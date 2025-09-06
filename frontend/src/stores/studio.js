import { defineStore } from "pinia";
import { StudioService } from "@/../bindings/clustta/services";
import { useProjectStore } from '@/stores/projects';
// import axios from "axios";
// import studio_service from "@/services/studio_service";
// import { useNotificationStore } from "./notifications";
// import { useAuthStore } from "./auth";

export const useStudioStore = defineStore("studio", {
  state: () => ({
    studios: [],
    allStudios: [],
    selectedStudioProjects: [],
    allStudiosLoaded: false,
    studiosLoaded: false,
    projectsLoaded: false,
    selectedStudio: null,
    selectedStudioUsers: [],
    studioUsers: []
  }),
  getters: {
    getStudiosNames: (state) => {
      console.log(state.studios);
      if (state.studios) {
        return state.studios.map((studio) => studio.name);
      }
      return [];
    },
    getSelectedStudioName: (state) => {
      if (state.selectedStudio) {
        return state.selectedStudio.name;
      }
      return "";
    },
    getSelectedStudioUsers: (state) => {
      return state.selectedStudioUsers;
    },
  },
  actions: {
    
    async loadStudios() {
      const notificationStore = useNotificationStore();
      try {
        this.studiosLoaded = false;
        const studiosResponse = await axios.get("/api/person/studios");
        let allStudios = studiosResponse.data;
        this.studios = allStudios.filter((studio) => studio.url);
        console.log(this.studios)
        // sort studios by name
        this.studios.sort((a, b) => a.name.localeCompare(b.name));

        // Use Promise.all to handle concurrent async operations
        const studiosWithUsers = await Promise.all(
          this.studios.map(async (studio) => {
            try {
              const studioStatus = await studio_service.getStudioStatus(
                studio.url
              );
              const usersResponse = await studio_service.getStudioUsers(
                studio.id
              );

              // sort users by name
              usersResponse.sort((a, b) => {
                let fullNameA = `${a.first_name} ${a.last_name}`.toLowerCase();
                let fullNameB = `${b.first_name} ${b.last_name}`.toLowerCase();
                return fullNameA.localeCompare(fullNameB);
              });

              return {
                ...studio,
                status: studioStatus,
                photo: "/icons/studio.svg",
                role: "admin",
                users: usersResponse,
              };
            } catch (error) {
              console.error(
                `Error fetching users for studio ${studio.id}:`,
                error
              );
              return {
                ...studio,
                status: "offline",
                photo: "/icons/studio.svg",
                role: "admin",
                users: [],
              };
            }
          })
        );

        // Update studios with user information
        this.studios = studiosWithUsers;

        // Select first studio if available
        if (this.studios.length > 0) {
          this.selectedStudio = this.studios[0];
        }
        this.studiosLoaded = true;
      } catch (error) {
        this.studiosLoaded = true;
        console.error("Error loading studios:", error);
        notificationStore.errorNotification("Error loading studios:", error);
      }
    },
    async getAllStudios() {
      const notificationStore = useNotificationStore();
      try {
        this.allStudiosLoaded = false;
        const studiosResponse = await axios.get("/api/person/all-studios");
        this.allStudios = studiosResponse.data;
        // sort studios by name
        this.allStudios.sort((a, b) => a.name.localeCompare(b.name));

        // Use Promise.all to handle concurrent async operations
        const studiosWithUsers = await Promise.all(
          this.allStudios.map(async (studio) => {
            try {
              const studioStatus = await studio_service.getStudioStatus(
                studio.url
              );
              const usersResponse = await studio_service.getStudioUsers(
                studio.id
              );

              // sort users by name
              usersResponse.sort((a, b) => {
                let fullNameA = `${a.first_name} ${a.last_name}`.toLowerCase();
                let fullNameB = `${b.first_name} ${b.last_name}`.toLowerCase();
                return fullNameA.localeCompare(fullNameB);
              });

              return {
                ...studio,
                status: studioStatus,
                photo: "/icons/studio.svg",
                role: "admin",
                users: usersResponse,
              };
            } catch (error) {
              console.error(
                `Error fetching users for studio ${studio.name}:`,
                error
              );
              return {
                ...studio,
                status: "offline",
                photo: "/icons/studio.svg",
                role: "admin",
                users: [],
              };
            }
          })
        );

        // Update studios with user information
        this.allStudios = studiosWithUsers;

        // Select first studio if available
        // if (this.allStudios.length > 0) {
        //   this.selectedStudio = this.allStudios[0];
        // }
        this.allStudiosLoaded = true;
      } catch (error) {
        this.allStudiosLoaded = true;
        console.error("Error loading studios:", error);
        notificationStore.errorNotification("Error loading studios:", error);
      }
    },
    async selectStudio(studio) {
      const authStore = useAuthStore();

      this.selectedStudio = studio;
      this.selectedStudioUsers = studio.users;

      this.loadStudioProjects(studio.url);

      // get all users photos in background
      studio.users.forEach((user) => {
        if (!user.photo) {
          authStore
            .getUserPhoto(user.id)
            .then((response) => {
              user.photo = response;
            })
            .catch((err) => {
              console.error(err);
              throw new Error("Error fetching user photo");
            });
        }
      });

      // studio_service
      //   .getStudioUsers(studio.id)
      //   .then((response) => {
      //     this.selectedStudio = studio;
      //     this.selectedStudioUsers = response;
      //   })
      //   .catch((error) => {
      //     console.log(error);
      //   });
    },

    async getStudioUsers() {
      const projectStore = useProjectStore();
      await StudioService.GetStudioUsers(projectStore.selectedStudio?.id).then( async( users) => {
          users.forEach((user) => {
              if(user.photo){
                  user.photo = "data:image/png;base64," + user.photo
              }
            })
        this.studioUsers = this.sortAlphabetically(users);
      }) 
    },

    sortAlphabetically(data){
      return data.sort((a, b) => a.first_name.localeCompare(b.first_name));
    },

    async loadStudioProjects(studioUrl) {
      this.projectsLoaded = false;
      studio_service
        .getStudioProjects(studioUrl)
        .then((response) => {
          if (typeof response === "object") {
            // console.log(response)
            this.selectedStudioProjects = response;
            this.projectsLoaded = true;
          }
        })
        .catch((error) => {
          this.projectsLoaded = true;
          console.log(error);
        });
    },
    async leaveStudio(studio) {
      this.studios = this.studios.filter((item) => item.id !== studio.id);
      this.selectedStudio = this.studios[0];
    },

  },
});
