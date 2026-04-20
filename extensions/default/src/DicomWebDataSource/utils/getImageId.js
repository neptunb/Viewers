import getWADORSImageId from './getWADORSImageId';

function buildInstanceWadoUrl(config, instance) {
  const { StudyInstanceUID, SeriesInstanceUID, SOPInstanceUID } = instance;
  const params = [];

  params.push('requestType=WADO');
  params.push(`studyUID=${StudyInstanceUID}`);
  params.push(`seriesUID=${SeriesInstanceUID}`);
  params.push(`objectUID=${SOPInstanceUID}`);
  params.push('contentType=application/dicom');
  params.push('transferSyntax=*');

  const paramString = params.join('&');

  return `${config.wadoUriRoot}?${paramString}`;
}

/**
 * Obtain an imageId for Cornerstone from an image instance
 *
 * @param instance
 * @param frame
 * @param thumbnail
 * @returns {string} The imageId to be used by Cornerstone
 */
export default function getImageId({ instance, frame, config, thumbnail = false }) {
  if (!instance) {
    return;
  }

  if (instance.imageId && frame === undefined) {
    return instance.imageId;
  }

  if (instance.url) {
    // Multi-frame support: append `frame=N` to the instance URL so each frame
    // gets a unique imageId. Without this, cornerstone caches all frames under
    // the same imageId and only the first frame is ever displayed (NM, XA,
    // enhanced CT/MR multi-frame objects were all affected).
    if (frame !== undefined) {
      const separator = instance.url.includes('?') ? '&' : '?';
      return `${instance.url}${separator}frame=${frame}`;
    }
    return instance.url;
  }

  const renderingAttr = thumbnail ? 'thumbnailRendering' : 'imageRendering';

  if (!config[renderingAttr] || config[renderingAttr] === 'wadouri') {
    const wadouri = buildInstanceWadoUrl(config, instance);

    let imageId = 'dicomweb:' + wadouri;
    if (frame !== undefined) {
      imageId += '&frame=' + frame;
    }

    return imageId;
  } else {
    return getWADORSImageId(instance, config, frame); // WADO-RS Retrieve Frame
  }
}
