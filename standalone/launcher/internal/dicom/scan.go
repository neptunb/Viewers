package dicom

import (
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	godicom "github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// BuildManifest walks root, parses every DICOM file it can read, and groups
// instances into study/series to match the OHIF DICOM JSON Model.
//
// Each instance's URL is generated as a relative path the Go HTTP server
// serves under /study/ (see internal/server). The dicomjson prefix
// ("dicomweb:" / wadouri) is added automatically by OHIF's DicomJSONDataSource
// when imageRendering is set to wadouri (default for dicomjson).
func BuildManifest(root string) (*Manifest, error) {
	type parsed struct {
		meta InstanceMetadata
		rel  string
	}

	var instances []parsed

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !looksLikeDICOM(path) {
			return nil
		}

		meta, ok := parseInstance(path)
		if !ok {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		instances = append(instances, parsed{meta: meta, rel: rel})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	studies := make(map[string]*Study)
	seriesIndex := make(map[string]*Series) // key: studyUID|seriesUID

	for _, p := range instances {
		meta := p.meta
		if meta.StudyInstanceUID == "" || meta.SeriesInstanceUID == "" || meta.SOPInstanceUID == "" {
			continue
		}

		st, ok := studies[meta.StudyInstanceUID]
		if !ok {
			st = &Study{
				StudyInstanceUID: meta.StudyInstanceUID,
				StudyDescription: meta.StudyDescription,
				StudyDate:        meta.StudyDate,
				StudyTime:        meta.StudyTime,
				AccessionNumber:  meta.AccessionNumber,
				PatientName:      meta.PatientName,
				PatientID:        meta.PatientID,
				PatientSex:       meta.PatientSex,
				PatientBirthDate: meta.PatientBirthDate,
			}
			studies[meta.StudyInstanceUID] = st
		}

		key := meta.StudyInstanceUID + "|" + meta.SeriesInstanceUID
		se, ok := seriesIndex[key]
		if !ok {
			se = &Series{
				SeriesInstanceUID: meta.SeriesInstanceUID,
				SeriesDescription: meta.SeriesDescription,
				SeriesNumber:      meta.SeriesNumber,
				SeriesDate:        meta.SeriesDate,
				SeriesTime:        meta.SeriesTime,
				Modality:          meta.Modality,
				BodyPartExamined:  meta.BodyPartExamined,
			}
			seriesIndex[key] = se
			st.Series = append(st.Series, se)
		}

		instanceURL := "/study/" + urlEscapePath(p.rel)
		se.Instances = append(se.Instances, &Instance{
			Metadata: meta,
			URL:      "dicomweb:" + instanceURL,
		})
		st.NumInstances++
	}

	if len(studies) == 0 {
		return &Manifest{}, nil
	}

	manifest := &Manifest{}
	for _, st := range studies {
		sort.SliceStable(st.Series, func(i, j int) bool {
			return st.Series[i].SeriesNumber < st.Series[j].SeriesNumber
		})
		modSet := make(map[string]struct{})
		for _, se := range st.Series {
			sort.SliceStable(se.Instances, func(i, j int) bool {
				return se.Instances[i].Metadata.InstanceNumber < se.Instances[j].Metadata.InstanceNumber
			})
			if se.Modality != "" {
				modSet[se.Modality] = struct{}{}
			}
		}
		mods := make([]string, 0, len(modSet))
		for m := range modSet {
			mods = append(mods, m)
		}
		sort.Strings(mods)
		st.Modalities = strings.Join(mods, "\\")
		manifest.Studies = append(manifest.Studies, st)
	}

	sort.SliceStable(manifest.Studies, func(i, j int) bool {
		return manifest.Studies[i].StudyInstanceUID < manifest.Studies[j].StudyInstanceUID
	})
	return manifest, nil
}

func looksLikeDICOM(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".dcm", ".dicom", ".ima":
		return true
	case "":
		return true // many PACS exports have no extension
	}
	return false
}

func parseInstance(path string) (InstanceMetadata, bool) {
	f, err := os.Open(path)
	if err != nil {
		return InstanceMetadata{}, false
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return InstanceMetadata{}, false
	}
	// Skip pixel data – we only need the header for the manifest.
	ds, err := godicom.Parse(f, fi.Size(), nil, godicom.SkipPixelData())
	if err != nil {
		return InstanceMetadata{}, false
	}

	meta := InstanceMetadata{}
	meta.StudyInstanceUID = str(ds, tag.StudyInstanceUID)
	meta.SeriesInstanceUID = str(ds, tag.SeriesInstanceUID)
	meta.SOPInstanceUID = str(ds, tag.SOPInstanceUID)
	if meta.StudyInstanceUID == "" || meta.SeriesInstanceUID == "" || meta.SOPInstanceUID == "" {
		return meta, false
	}
	meta.SOPClassUID = str(ds, tag.SOPClassUID)
	meta.Modality = str(ds, tag.Modality)
	meta.InstanceNumber = intVal(ds, tag.InstanceNumber)
	meta.SeriesNumber = intVal(ds, tag.SeriesNumber)
	meta.Rows = intVal(ds, tag.Rows)
	meta.Columns = intVal(ds, tag.Columns)
	meta.NumberOfFrames = intVal(ds, tag.NumberOfFrames)
	meta.BitsAllocated = intVal(ds, tag.BitsAllocated)
	meta.BitsStored = intVal(ds, tag.BitsStored)
	meta.HighBit = intVal(ds, tag.HighBit)
	meta.PixelRepresentation = intVal(ds, tag.PixelRepresentation)
	meta.SamplesPerPixel = intVal(ds, tag.SamplesPerPixel)
	meta.PhotometricInterpretation = str(ds, tag.PhotometricInterpretation)
	meta.PlanarConfiguration = intVal(ds, tag.PlanarConfiguration)
	meta.PixelSpacing = floats(ds, tag.PixelSpacing)
	meta.ImagerPixelSpacing = floats(ds, tag.ImagerPixelSpacing)
	meta.SliceThickness = floatVal(ds, tag.SliceThickness)
	meta.SliceLocation = floatVal(ds, tag.SliceLocation)
	meta.ImagePositionPatient = floats(ds, tag.ImagePositionPatient)
	meta.ImageOrientationPatient = floats(ds, tag.ImageOrientationPatient)
	meta.FrameOfReferenceUID = str(ds, tag.FrameOfReferenceUID)
	meta.RescaleIntercept = floatVal(ds, tag.RescaleIntercept)
	meta.RescaleSlope = floatVal(ds, tag.RescaleSlope)
	meta.RescaleType = str(ds, tag.RescaleType)
	meta.WindowCenter = floats(ds, tag.WindowCenter)
	meta.WindowWidth = floats(ds, tag.WindowWidth)
	meta.PatientName = str(ds, tag.PatientName)
	meta.PatientID = str(ds, tag.PatientID)
	meta.PatientSex = str(ds, tag.PatientSex)
	meta.PatientBirthDate = str(ds, tag.PatientBirthDate)
	meta.StudyDescription = str(ds, tag.StudyDescription)
	meta.StudyDate = str(ds, tag.StudyDate)
	meta.StudyTime = str(ds, tag.StudyTime)
	meta.AccessionNumber = str(ds, tag.AccessionNumber)
	meta.SeriesDescription = str(ds, tag.SeriesDescription)
	meta.SeriesDate = str(ds, tag.SeriesDate)
	meta.SeriesTime = str(ds, tag.SeriesTime)
	meta.BodyPartExamined = str(ds, tag.BodyPartExamined)
	meta.Manufacturer = str(ds, tag.Manufacturer)
	meta.ManufacturerModelName = str(ds, tag.ManufacturerModelName)
	meta.SpacingBetweenSlices = floatVal(ds, tag.SpacingBetweenSlices)
	meta.AcquisitionNumber = intVal(ds, tag.AcquisitionNumber)
	meta.StationName = str(ds, tag.StationName)
	meta.KVP = floatVal(ds, tag.KVP)
	meta.TransferSyntaxUID = str(ds, tag.TransferSyntaxUID)
	return meta, true
}

func str(ds godicom.Dataset, t tag.Tag) string {
	el, err := ds.FindElementByTag(t)
	if err != nil {
		return ""
	}
	v := el.Value
	if v == nil {
		return ""
	}
	switch vv := v.GetValue().(type) {
	case []string:
		if len(vv) == 0 {
			return ""
		}
		return strings.TrimSpace(vv[0])
	case string:
		return strings.TrimSpace(vv)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", vv))
	}
}

// intVal supports every numeric VR the suyashkumar/dicom library may return
// (US/SS/UL/SL come back as []int, IS/DS as []string). Required because
// Rows/Columns/BitsAllocated etc. are US and would otherwise serialize as 0,
// breaking cornerstone-dicom-image-loader pixel decoding.
func intVal(ds godicom.Dataset, t tag.Tag) int {
	el, err := ds.FindElementByTag(t)
	if err != nil || el.Value == nil {
		return 0
	}
	switch v := el.Value.GetValue().(type) {
	case []int:
		if len(v) > 0 {
			return v[0]
		}
	case int:
		return v
	case []string:
		if len(v) > 0 {
			if n, err := strconv.Atoi(strings.TrimSpace(v[0])); err == nil {
				return n
			}
		}
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return 0
}

func floatVal(ds godicom.Dataset, t tag.Tag) float64 {
	if fs := floats(ds, t); len(fs) > 0 {
		return fs[0]
	}
	return 0
}

// floats supports FL/FD ([]float64), US/UL/SS/SL ([]int) and DS/IS ([]string).
func floats(ds godicom.Dataset, t tag.Tag) []float64 {
	el, err := ds.FindElementByTag(t)
	if err != nil || el.Value == nil {
		return nil
	}
	switch v := el.Value.GetValue().(type) {
	case []float64:
		if len(v) == 0 {
			return nil
		}
		out := make([]float64, len(v))
		copy(out, v)
		return out
	case []int:
		if len(v) == 0 {
			return nil
		}
		out := make([]float64, len(v))
		for i, n := range v {
			out[i] = float64(n)
		}
		return out
	case []string:
		out := make([]float64, 0, len(v))
		for _, s := range v {
			f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil {
				continue
			}
			out = append(out, f)
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}
	return nil
}

// urlEscapePath escapes each path segment while keeping '/' separators so the
// instance URL is valid when the filename contains spaces or UID-like chars.
func urlEscapePath(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		segs[i] = url.PathEscape(s)
	}
	return strings.Join(segs, "/")
}

var _ = log.Ltime // retain log import for future diagnostics
