export const projectDataChangedEvent = 'project-data-changed';

export const isActiveProjectChange = (change, activeProject) => {
  if (!change || !activeProject) return false;
  return change.project_id === activeProject.id
    && change.project_uri === activeProject.uri;
};
