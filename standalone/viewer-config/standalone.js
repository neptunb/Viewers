/**
 * Standalone OHIF Viewer config.
 *
 * Served by the Go launcher (macos_view / linux_view / windows_view.exe).
 * The launcher:
 *   - Scans ./study/ for DICOM files
 *   - Emits /study.json (DICOM JSON Model)
 *   - Serves DICOM files at /study/<relative path>
 *   - Redirects / to /index.html?datasources=dicomjson&url=/study.json&StudyInstanceUIDs=<uid>
 *
 * @type {AppTypes.Config}
 */
window.config = {
  routerBasename: '/',
  whiteLabeling: {},
  extensions: [],
  modes: [],
  customizationService: [],
  showStudyList: false,
  maxNumberOfWebWorkers: 4,
  showWarningMessageForCrossOrigin: false,
  showCPUFallbackMessage: true,
  showLoadingIndicator: true,
  strictZSpacingForVolumeViewport: true,
  defaultDataSourceName: 'dicomjson',
  dataSources: [
    {
      namespace: '@ohif/extension-default.dataSourcesModule.dicomjson',
      sourceName: 'dicomjson',
      configuration: {
        friendlyName: 'Local Study (DICOM JSON)',
        name: 'json',
      },
    },
    {
      namespace: '@ohif/extension-default.dataSourcesModule.dicomlocal',
      sourceName: 'dicomlocal',
      configuration: {
        friendlyName: 'Drag & Drop Local Files',
      },
    },
  ],
  httpErrorHandler: error => {
    console.warn('[standalone] HTTP error:', error?.status, error);
  },
};
