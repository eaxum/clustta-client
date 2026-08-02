import { useI18n } from 'vue-i18n';
import { refreshEntitlements } from '@/lib/sync';
import { FSService, ProjectService } from '@/services';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useMenu } from '@/stores/menu';
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useStudioStore } from '@/stores/studio';
import { useTrayStates } from '@/stores/TrayStates';

export const useProjectRemovalActions = () => {
  const menu = useMenu();
  const modals = useDesktopModalStore();
  const notificationStore = useNotificationStore();
  const projectStore = useProjectStore();
  const studioStore = useStudioStore();
  const trayStates = useTrayStates();
  const { t } = useI18n();

  const removeProject = async ({ deleteWorkingFiles } = {}) => {
    const project = projectStore.getActiveProject;
    await FSService.DeleteFile(project.uri);

    if (deleteWorkingFiles && project.working_directory) {
      await FSService.DeleteFolder(project.working_directory);
    }

    projectStore.removeProjectFromList(project.uri);
  };

  const deleteRemoteProject = async ({ deleteWorkingFiles } = {}) => {
    const project = projectStore.getActiveProject;
    await ProjectService.DeleteRemoteProject(
      projectStore.getActiveProjectUrl,
      projectStore.selectedStudio.name,
    );

    if (project.uri && project.is_downloaded) {
      await FSService.DeleteFile(project.uri);
    }
    if (deleteWorkingFiles && project.working_directory) {
      await FSService.DeleteFolder(project.working_directory);
    }

    projectStore.removeProjectFromList(project.uri, { force: true });
    refreshEntitlements();
    notificationStore.addNotification(
      t('notifications.projectDeleted'),
      t('notifications.projectDeletedDesc', { name: project.name }),
      'success',
      false,
    );
  };

  const configureWorkingFilesToggle = () => {
    trayStates.dangerousActionShowToggle = true;
    trayStates.dangerousActionToggleLabel = t('modals.confirmDangerousAction.deleteWorkingFiles');
    trayStates.dangerousActionToggleOffHint = t('modals.confirmDangerousAction.deleteWorkingFilesOff');
    trayStates.dangerousActionToggleOnHint = t('modals.confirmDangerousAction.deleteWorkingFilesOn');
  };

  const prepRemoveProject = () => {
    const project = projectStore.getActiveProject;
    trayStates.dangerousActionTitle = t('menus.removeProjectTitle', { name: project.name });
    trayStates.dangerousActionMessage = `${t('confirmations.deleteProjectLocal')} ${t('confirmations.deleteProjectTeamSuffix')}`;
    trayStates.dangerousActionIcon = 'minus-circle';
    trayStates.dangerousActionConfirmText = '';
    trayStates.dangerousActionShowInput = false;
    trayStates.dangerousActionFunction = removeProject;
    configureWorkingFilesToggle();
    modals.setModalVisibility('confirmDangerousActionModal', true);
    menu.hideContextMenu();
  };

  const prepDeleteProject = () => {
    const project = projectStore.getActiveProject;
    const isPersonal = projectStore.selectedStudio?.name === 'Personal';
    const canDeleteRemote = project.has_remote && studioStore.canManageProject;
    const localMessage = `${t('confirmations.deleteProjectLocal')} ${isPersonal
      ? t('confirmations.deleteProjectPersonalSuffix')
      : t('confirmations.deleteProjectTeamSuffix')}`;

    trayStates.dangerousActionTitle = t('menus.deleteProjectTitle', { name: project.name });
    trayStates.dangerousActionMessage = canDeleteRemote
      ? t('confirmations.deleteRemoteProject', { name: project.name })
      : localMessage;
    trayStates.dangerousActionIcon = 'trash';
    trayStates.dangerousActionConfirmText = project.name;
    trayStates.dangerousActionShowInput = true;
    trayStates.dangerousActionFunction = canDeleteRemote ? deleteRemoteProject : removeProject;
    configureWorkingFilesToggle();
    modals.setModalVisibility('confirmDangerousActionModal', true);
    menu.hideContextMenu();
  };

  return { prepDeleteProject, prepRemoveProject };
};
