import { defineStore } from "pinia";
import { UserService } from "@/../bindings/clustta/services";
import { useProjectStore } from "./projects";

export const useUserStore = defineStore("users", {
  state: () => ({
    user: null,
    users: [],
    users_index: {},
    roles: [],
    isUserAuthenticated: false,
    userCanCreateProject: false,
    selectedUser: null,
    selectedRole: null,
  }),
  getters: {
    getProjectCollaborators: (state) => {
      return state.users;
    },
    getUserAuthentication: (state) => {
      return state.isUserAuthenticated;
    },
    getProjectRoles: (state) => {
      return state.roles;
    },
    getRolesNames: (state) => {
      return state.roles.map((obj) => obj.name);
    },
    getSelectedUser: (state) => {
      return state.selectedUser;
    },
    getUser: (state) => {
      return state.user;
    },
  },
  actions: {
    getUserData(userId) {
      let user = this.users.find((user) => user.id === userId);
      return user;
    },
    selectUser(userId) {
      let user = this.users.find((user) => user.id === userId);
      this.selectedUser = user;
    },
    userProfileColor(uuid) {
      const parts = uuid.split("-");
      const hexCode = parts[0];
      const rgbHex = hexCode.length > 6 ? hexCode.substring(0, 6) : hexCode;
      const paddedHex = rgbHex.padEnd(6, '0');
      return '#' + paddedHex;

    },
    userProfilePhoto(userId) {
      let user = this.users.find((user) => user.id === userId);
      return user?.photo;
    },
    canDo(action) {
      if (!this.user) {
        return false;
      }
      if (this.user.role) {
        return this.user.role[action];
      } else {
        return false;
      }
    },
    async reloadUsers() {
      const projectStore = useProjectStore();
      let user = this.user;
      this.users = await UserService.GetUsers(projectStore.activeProject?.uri);
      for (let i = 0; i < this.users.length; i++) {
        if (this.users[i].photo) {
          this.users[i].photo = "data:image/png;base64," + this.users[i].photo;
        }
      }
      this.user = this.getUserData(user.id);
      let roles = await UserService.GetRoles(projectStore.activeProject.uri);
      this.roles = roles;
      this.rebuildUsersIndex();
    },
    async rebuildUsersIndex() {
      let userIndex = {};
      for (let i = 0; i < this.users.length; i++) {
        userIndex[this.users[i].id] = i;
      }
      this.users_index = userIndex;
    },
  },
});
