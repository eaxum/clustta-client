import { ProjectService } from "@/../bindings/clustta/services";
import { useNotificationStore } from "@/stores/notifications";
import { useProjectStore } from "@/stores/projects";
import { useTrayStates } from "@/stores/TrayStates";

export async function addIgnoredItem(item) {
    const projectStore = useProjectStore();
    const notificationStore = useNotificationStore();

    let ignoreList = projectStore.activeProject.ignore_list;
    ignoreList.push(item);
    let projectUri;
    if (projectStore.activeProject.has_remote) {
        projectUri = projectStore.getActiveProjectUrl;
    } else {
        projectUri = projectStore.activeProject.uri;
    }
    await ProjectService.SetIgnoreList(
        projectUri,
        projectStore.selectedStudio.name,
        ignoreList
    )
        .then(async (response) => {
        let project = projectStore.activeProject;
        let projectIndex = projectStore.projects.findIndex(
            (p) => p.id === project.id
        );
        projectStore.projects[projectIndex].ignore_list = ignoreList;
        projectStore.activeProject.ignore_list = ignoreList;
        // notificationStore.addNotification('Ignore list updated', 'success');
        await projectStore.refreshUntrackedItems();
        })
        .catch((error) => {
        console.log(error);
        notificationStore.addNotification(
            "Failed to update ignore list",
            "error"
        );
        });
}
