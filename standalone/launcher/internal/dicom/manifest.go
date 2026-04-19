package dicom

// Manifest mirrors the DICOM JSON Model consumed by OHIF's dicomjson data source.
// The structure matches what platform DicomJSONDataSource expects:
//   { studies: [ { StudyInstanceUID, ..., series: [ { SeriesInstanceUID, ..., instances: [ { metadata, url } ] } ] } ] }
type Manifest struct {
	Studies []*Study `json:"studies"`
}

type Study struct {
	StudyInstanceUID string    `json:"StudyInstanceUID"`
	StudyDescription string    `json:"StudyDescription,omitempty"`
	StudyDate        string    `json:"StudyDate,omitempty"`
	StudyTime        string    `json:"StudyTime,omitempty"`
	AccessionNumber  string    `json:"AccessionNumber,omitempty"`
	PatientName      string    `json:"PatientName,omitempty"`
	PatientID        string    `json:"PatientID,omitempty"`
	PatientSex       string    `json:"PatientSex,omitempty"`
	PatientBirthDate string    `json:"PatientBirthDate,omitempty"`
	Modalities       string    `json:"Modalities,omitempty"`
	NumInstances     int       `json:"NumInstances"`
	Series           []*Series `json:"series"`
}

type Series struct {
	SeriesInstanceUID string      `json:"SeriesInstanceUID"`
	SeriesDescription string      `json:"SeriesDescription,omitempty"`
	SeriesNumber      int         `json:"SeriesNumber,omitempty"`
	SeriesDate        string      `json:"SeriesDate,omitempty"`
	SeriesTime        string      `json:"SeriesTime,omitempty"`
	Modality          string      `json:"Modality,omitempty"`
	BodyPartExamined  string      `json:"BodyPartExamined,omitempty"`
	Instances         []*Instance `json:"instances"`
}

type Instance struct {
	Metadata InstanceMetadata `json:"metadata"`
	URL      string           `json:"url"`
}

// InstanceMetadata is the naturalized DICOM metadata OHIF's dicomjson
// data source ingests verbatim into its DicomMetadataStore.
// Fields left as omitempty when absent to keep manifests compact.
type InstanceMetadata struct {
	SOPInstanceUID             string    `json:"SOPInstanceUID"`
	SOPClassUID                string    `json:"SOPClassUID,omitempty"`
	SeriesInstanceUID          string    `json:"SeriesInstanceUID,omitempty"`
	StudyInstanceUID           string    `json:"StudyInstanceUID,omitempty"`
	InstanceNumber             int       `json:"InstanceNumber,omitempty"`
	SeriesNumber               int       `json:"SeriesNumber,omitempty"`
	Modality                   string    `json:"Modality,omitempty"`
	Rows                       int       `json:"Rows,omitempty"`
	Columns                    int       `json:"Columns,omitempty"`
	NumberOfFrames             int       `json:"NumberOfFrames,omitempty"`
	BitsAllocated              int       `json:"BitsAllocated,omitempty"`
	BitsStored                 int       `json:"BitsStored,omitempty"`
	HighBit                    int       `json:"HighBit,omitempty"`
	PixelRepresentation        int       `json:"PixelRepresentation,omitempty"`
	SamplesPerPixel            int       `json:"SamplesPerPixel,omitempty"`
	PhotometricInterpretation  string    `json:"PhotometricInterpretation,omitempty"`
	PlanarConfiguration        int       `json:"PlanarConfiguration,omitempty"`
	PixelSpacing               []float64 `json:"PixelSpacing,omitempty"`
	ImagerPixelSpacing         []float64 `json:"ImagerPixelSpacing,omitempty"`
	SliceThickness             float64   `json:"SliceThickness,omitempty"`
	SliceLocation              float64   `json:"SliceLocation,omitempty"`
	ImagePositionPatient       []float64 `json:"ImagePositionPatient,omitempty"`
	ImageOrientationPatient    []float64 `json:"ImageOrientationPatient,omitempty"`
	FrameOfReferenceUID        string    `json:"FrameOfReferenceUID,omitempty"`
	RescaleIntercept           float64   `json:"RescaleIntercept,omitempty"`
	RescaleSlope               float64   `json:"RescaleSlope,omitempty"`
	RescaleType                string    `json:"RescaleType,omitempty"`
	WindowCenter               []float64 `json:"WindowCenter,omitempty"`
	WindowWidth                []float64 `json:"WindowWidth,omitempty"`
	PatientName                string    `json:"PatientName,omitempty"`
	PatientID                  string    `json:"PatientID,omitempty"`
	PatientSex                 string    `json:"PatientSex,omitempty"`
	PatientBirthDate           string    `json:"PatientBirthDate,omitempty"`
	StudyInstanceUIDRef        string    `json:"-"`
	StudyDescription           string    `json:"StudyDescription,omitempty"`
	StudyDate                  string    `json:"StudyDate,omitempty"`
	StudyTime                  string    `json:"StudyTime,omitempty"`
	AccessionNumber            string    `json:"AccessionNumber,omitempty"`
	SeriesDescription          string    `json:"SeriesDescription,omitempty"`
	SeriesDate                 string    `json:"SeriesDate,omitempty"`
	SeriesTime                 string    `json:"SeriesTime,omitempty"`
	BodyPartExamined           string    `json:"BodyPartExamined,omitempty"`
	Manufacturer               string    `json:"Manufacturer,omitempty"`
	ManufacturerModelName      string    `json:"ManufacturerModelName,omitempty"`
	SpacingBetweenSlices       float64   `json:"SpacingBetweenSlices,omitempty"`
	SOPInstanceUIDSanityChk    string    `json:"-"`
	AcquisitionNumber          int       `json:"AcquisitionNumber,omitempty"`
	StationName                string    `json:"StationName,omitempty"`
	KVP                        float64   `json:"KVP,omitempty"`
	TransferSyntaxUID          string    `json:"TransferSyntaxUID,omitempty"`
}

// CountSeries returns total series across all studies in the manifest.
func (m *Manifest) CountSeries() int {
	n := 0
	for _, s := range m.Studies {
		n += len(s.Series)
	}
	return n
}

// CountInstances returns total instances across all studies/series.
func (m *Manifest) CountInstances() int {
	n := 0
	for _, s := range m.Studies {
		for _, se := range s.Series {
			n += len(se.Instances)
		}
	}
	return n
}
