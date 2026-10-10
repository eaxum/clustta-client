import UserAppearance from '@/instances/desktop/settings/UserAppearance.vue';
import UserBehavior from '@/instances/desktop/settings/UserBehavior.vue';
import UserExperimental from '@/instances/desktop/settings/UserExperimental.vue';
import UserHelpAbout from '@/instances/desktop/settings/UserHelpAbout.vue';
import Directories from '@/instances/desktop/settings/Directories.vue';
import ProjectTemplates from '@/instances/desktop/settings/ProjectTemplates.vue';
import UserIntegrations from '@/instances/desktop/settings/UserIntegrations.vue';
import Studio from '@/instances/desktop/settings/Studio.vue';
import ProjectStorage from '@/instances/desktop/settings/ProjectStorage.vue';
import StudioCollaborators from '@/instances/desktop/settings/StudioCollaborators.vue';
import StudioIntegrations from '@/instances/desktop/settings/StudioIntegrations.vue';
import StudioAuditLogs from '@/instances/desktop/settings/StudioAuditLogs.vue';
import Collaborators from '@/instances/desktop/settings/Collaborators.vue';
import Roles from '@/instances/desktop/settings/Roles.vue';
import Templates from '@/instances/desktop/settings/Templates.vue';
import WorkflowTemplates from '@/instances/desktop/settings/WorkflowTemplates.vue';
import AssetTypes from '@/instances/desktop/settings/AssetTypes.vue';
import CollectionTypes from '@/instances/desktop/settings/CollectionTypes.vue';
import Tags from '@/instances/desktop/settings/Tags.vue';
import IgnoreList from '@/instances/desktop/settings/IgnoreList.vue';
import Hooks from '@/instances/desktop/settings/Hooks.vue';
import Advanced from '@/instances/desktop/settings/Advanced.vue';
import ProjectIntegrations from '@/instances/desktop/settings/ProjectIntegrations.vue';
import { canAccessProjectSettingsTab } from '@/lib/permissions';

export const userSettingsGroups = [
  { id: 'preferences', nameKey: 'settings.navigation.preferences' },
  { id: 'files', nameKey: 'settings.navigation.filesAndProjects' },
  { id: 'connections', nameKey: 'settings.navigation.connections' },
  { id: 'system', nameKey: 'settings.navigation.system' },
];

export const userSettingsPages = [
  { id: 'appearance', nameKey: 'settings.appearance', icon: 'palette', group: 'preferences', component: UserAppearance },
  { id: 'behavior', nameKey: 'settings.behaviour', icon: 'cog', group: 'preferences', component: UserBehavior },
  { id: 'directories', nameKey: 'settings.directories', icon: 'explorer', group: 'files', component: Directories },
  { id: 'projecttemplates', nameKey: 'settings.projectTemplates', icon: 'briefcase', group: 'files', component: ProjectTemplates },
  { id: 'userintegrations', nameKey: 'settings.integrations', icon: 'plug', group: 'connections', component: UserIntegrations },
  { id: 'experimental', nameKey: 'settings.experimentalFeatures', icon: 'skull', group: 'system', component: UserExperimental },
  { id: 'helpabout', nameKey: 'settings.resourcesSupport', icon: 'help', group: 'system', component: UserHelpAbout },
];

export const studioSettingsGroups = [
  { id: 'studio', nameKey: 'settings.navigation.studio' },
  { id: 'management', nameKey: 'settings.navigation.management' },
  { id: 'connections', nameKey: 'settings.navigation.connections' },
];

export const studioSettingsPages = [
  { id: 'studio', nameKey: 'settings.studio', icon: 'stall', group: 'studio', component: Studio },
  { id: 'studioprojects', nameKey: 'settings.projectStorage', icon: 'briefcase', group: 'management', component: ProjectStorage },
  { id: 'studiocollaborators', nameKey: 'settings.studioCollaborators', icon: 'person', group: 'management', component: StudioCollaborators },
  { id: 'auditlogs', nameKey: 'settings.auditLogs', icon: 'file', group: 'management', badgeKey: 'settings.comingSoon', component: StudioAuditLogs },
  { id: 'studiointegrations', nameKey: 'settings.studioIntegrations', icon: 'plug', group: 'connections', component: StudioIntegrations },
];

export const getAvailableStudioSettingsPages = ({ canCollaborate, isCloudHosted }) => {
  return studioSettingsPages.filter((page) => {
    if (page.id === 'studioprojects') return !isCloudHosted;
    if (page.id === 'studiocollaborators') return canCollaborate;
    return true;
  });
};

export const projectSettingsGroups = [
  { id: 'people', nameKey: 'settings.navigation.people' },
  { id: 'configuration', nameKey: 'settings.navigation.configuration' },
  { id: 'files', nameKey: 'settings.navigation.filesAndAutomation' },
  { id: 'connections', nameKey: 'settings.navigation.connections' },
  { id: 'system', nameKey: 'settings.navigation.system' },
];

export const projectSettingsPages = [
  { id: 'collaborators', nameKey: 'settings.collaborators', icon: 'person', group: 'people', component: Collaborators },
  { id: 'roles', nameKey: 'settings.roles', icon: 'scale', group: 'people', component: Roles },
  { id: 'templates', nameKey: 'settings.templates', icon: 'file', group: 'configuration', component: Templates },
  { id: 'workflows', nameKey: 'settings.workflows', icon: 'workflow-arrow', group: 'configuration', component: WorkflowTemplates },
  { id: 'assettypes', nameKey: 'settings.assetTypes', icon: 'brush', group: 'configuration', component: AssetTypes },
  { id: 'collectiontypes', nameKey: 'settings.collectionTypes', icon: 'folder', group: 'configuration', component: CollectionTypes },
  { id: 'tags', nameKey: 'settings.tags', icon: 'tag', group: 'configuration', component: Tags },
  { id: 'ignorelist', nameKey: 'settings.ignoreList', icon: 'file-watch', group: 'files', component: IgnoreList },
  { id: 'hooks', nameKey: 'settings.launchHooks', icon: 'hook', group: 'files', component: Hooks },
  { id: 'integrations', nameKey: 'settings.integrations', icon: 'plug', group: 'connections', component: ProjectIntegrations },
  { id: 'advanced', nameKey: 'settings.advanced', icon: 'skull', group: 'system', component: Advanced },
];

export const getAvailableProjectSettingsPages = ({
  canCollaborate,
  hasCustomRoles,
  hasIntegrations,
  isRemoteProject,
}) => {
  return projectSettingsPages.filter((page) => {
    if (!canAccessProjectSettingsTab(page.id)) return false;
    if (page.id === 'collaborators') return isRemoteProject && canCollaborate;
    if (page.id === 'roles') return isRemoteProject && hasCustomRoles;
    if (page.id === 'integrations') return hasIntegrations;
    return true;
  });
};
