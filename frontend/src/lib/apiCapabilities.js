export const VERSIONED_DEPENDENCIES = 'versioned_dependencies';
export const PROJECT_PERMISSIONS = 'project_permissions';
export const CURRENT_PROJECT_SCHEMA = '2.2';

const LEGACY_API_VERSION = '1';
const CURRENT_API_VERSION = '2';
const CLIENT_API_VERSIONS = [LEGACY_API_VERSION, CURRENT_API_VERSION];

export function currentAPIInfo() {
  return {
    default_version: LEGACY_API_VERSION,
    supported_versions: [...CLIENT_API_VERSIONS],
    capabilities_by_version: {
      [LEGACY_API_VERSION]: [],
      [CURRENT_API_VERSION]: [VERSIONED_DEPENDENCIES, PROJECT_PERMISSIONS],
    },
  };
}

export function negotiateStudioAPI(apiInfo) {
  const normalized = normalizeAPIInfo(apiInfo);
  const supported = new Set(normalized.supported_versions);
  const version = [...CLIENT_API_VERSIONS].reverse().find(candidate => supported.has(candidate));

  return {
    info: normalized,
    version: version || null,
    capabilities: version
      ? [...(normalized.capabilities_by_version[version] || [])]
      : [],
  };
}

export function normalizeAPIInfo(apiInfo) {
  if (!apiInfo || !Array.isArray(apiInfo.supported_versions)) {
    return {
      default_version: LEGACY_API_VERSION,
      supported_versions: [LEGACY_API_VERSION],
      capabilities_by_version: { [LEGACY_API_VERSION]: [] },
    };
  }

  const supportedVersions = apiInfo.supported_versions.filter(
    version => typeof version === 'string' && /^\d+$/.test(version),
  );
  const capabilitiesByVersion = {};
  for (const version of supportedVersions) {
    const capabilities = apiInfo.capabilities_by_version?.[version];
    capabilitiesByVersion[version] = Array.isArray(capabilities)
      ? capabilities.filter(capability => typeof capability === 'string')
      : [];
  }

  return {
    default_version: typeof apiInfo.default_version === 'string'
      ? apiInfo.default_version
      : LEGACY_API_VERSION,
    supported_versions: supportedVersions,
    capabilities_by_version: capabilitiesByVersion,
  };
}

