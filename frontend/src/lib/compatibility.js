export const PROJECT_PROTOCOL = '1';
export const PROJECT_SCHEMA = '2.2';
export const COMPATIBILITY_ERROR = 'project_schema_unsupported';
export const COMPATIBILITY_CACHE_KEY = 'clustta.projectCompatibility';

export function projectCompatibilityProblem(project, cached) {
  if (project.is_offline) {
    if (cached?.problem) return cached.problem;
    const problem = compatibilityProblem(cached?.verified ? cached.contract : null);
    if (problem) return problem;
    if (project.is_downloaded && project.local_schema && project.local_schema !== cached.contract.project_schema) {
      return replicaProblem();
    }
    return null;
  }
  const problem = compatibilityProblem(project.compatibility);
  if (problem) return problem;
  if (project.is_downloaded && project.local_schema && project.local_schema !== project.compatibility.project_schema) {
    return replicaProblem();
  }
  return null;
}

export function projectAccessProblem(project, cached) {
  const remoteProblem = projectCompatibilityProblem(project, cached);
  if (!project.is_downloaded) return remoteProblem;
  if (project.local_schema && project.local_schema !== PROJECT_SCHEMA) return replicaProblem();
  if (project.is_offline && !cached?.verified && !cached?.problem) return remoteProblem;
  return null;
}

export function compatibilityProblem(contract) {
  if (!contract) {
    return { code: COMPATIBILITY_ERROR, required_update: 'server', message: 'Update the Studio/server to access this project.' };
  }
  if (!validProtocol(contract.protocol) || !validSchema(contract.schema) || !validSchema(contract.project_schema)) {
    return { code: COMPATIBILITY_ERROR, required_update: 'server', message: 'Update the Studio/server to access this project.' };
  }
  if (contract.protocol === PROJECT_PROTOCOL && contract.schema === PROJECT_SCHEMA && contract.project_schema === PROJECT_SCHEMA) return null;
  let update = 'client';
  if (contract.protocol !== PROJECT_PROTOCOL) {
    update = Number(PROJECT_PROTOCOL) > Number(contract.protocol) ? 'server' : 'client';
  } else if (contract.project_schema !== contract.schema) {
    update = 'server';
  } else if (compareSchema(PROJECT_SCHEMA, contract.schema) > 0) {
    update = 'server';
  }
  return {
    code: COMPATIBILITY_ERROR,
    required_update: update,
    message: update === 'server' ? 'Update the Studio/server to access this project.' : 'Update Clustta to access this project.',
  };
}

function replicaProblem() {
  return {
    code: COMPATIBILITY_ERROR,
    required_update: 'replica',
    message: 'This local replica uses a different project schema. Local changes have been preserved; update the replica before syncing.',
  };
}

function compareSchema(left, right) {
  const leftParts = left.split('.').map(Number);
  const rightParts = right.split('.').map(Number);
  if (leftParts.length !== 2 || rightParts.length !== 2 || [...leftParts, ...rightParts].some(Number.isNaN)) return 0;
  if (leftParts[0] !== rightParts[0]) return leftParts[0] - rightParts[0];
  return leftParts[1] - rightParts[1];
}

function validProtocol(value) {
  return /^[1-9]\d*$/.test(value);
}

function validSchema(value) {
  return /^(0|[1-9]\d*)\.(0|[1-9]\d*)$/.test(value);
}

export function parseCompatibilityError(error) {
  if (error?.code === COMPATIBILITY_ERROR) return error;
  const message = error?.message || String(error);
  const start = message.indexOf('{');
  if (start < 0) return null;
  try {
    const rejection = JSON.parse(message.slice(start));
    return rejection.code === COMPATIBILITY_ERROR ? rejection : null;
  } catch {
    return null;
  }
}

export function compatibilityCacheKey(userId, studioId, project) {
  return JSON.stringify([userId, studioId, project.remote, project.id]);
}
