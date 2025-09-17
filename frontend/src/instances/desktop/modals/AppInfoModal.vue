<template>

  <div class="modal-container" v-esc="closeModal">
    <div class="single-action-button close" @click="closeModal()">
      <img class="small-icons" src="/icons/close.svg">
    </div>
    <HeaderArea :title="title" :icon="'clustta'" :showSearch="showSearch" />

    <div class="general-container">
      <div class="version-info">
        <!-- <div> Clustta for Windows (64 bit) </div> -->
        <div> Version {{ clusttaVersion }} </div>
        <div> Copyright &copy; 2024 Clustta </div>
        <div> Clustta<sup>&reg;</sup> is a registered trademark of Eaxum LLC. </div>
      </div>

      <!-- <div class="pop-up-actions">
        <button class="button" @click="checkForUpdates()" v-stop-propagation>

          <div class="login-button-text">
            {{ isAwaitingResponse ? 'Checking for Updates' : 'Check for Updates' }}
          </div>

          <div v-if="isAwaitingResponse" class="login-button-icon loading-icon">
            <img src="/icons/loading.svg" />
          </div>

        </button>
      </div> -->




    </div>
  </div>
</template>

<script setup>
// imports
import { ref, onMounted } from 'vue';
import utils from "@/services/utils";

// store/state imports
import { useCollectionStore } from '@/stores/collections';
import { useDesktopModalStore } from '@/stores/desktopModals';

// components
import HeaderArea from '@/instances/common/components/HeaderArea.vue';

// vars
let title = 'Clustta';
let showSearch = false;

// refs
const entityName = ref('');
const clusttaVersion = ref("");
const isAwaitingResponse = ref(false);

// states/stores
const collectionStore = useCollectionStore();
const modals = useDesktopModalStore();

// methods
const closeModal = () => {
  modals.disableAllModals();
};

const checkForUpdates = async () => {
  isAwaitingResponse.value = true;
  const update = await check();
  if (update) {
    console.log(
      `found update ${update.version} from ${update.date} with notes ${update.body}`
    );
    let downloaded = 0;
    let contentLength = 0;
    // alternatively we could also call update.download() and update.install() separately
    await update.download((event) => {
      switch (event.event) {
        case 'Started':
          contentLength = event.data.contentLength;
          console.log(`started downloading ${event.data.contentLength} bytes`);
          break;
        case 'Progress':
          downloaded += event.data.chunkLength;
          console.log(`downloaded ${downloaded} from ${contentLength}`);
          break;
        case 'Finished':
          console.log('download finished');
          break;
      }
    });

    console.log('update installed');
  }
  isAwaitingResponse.value = false;
};

onMounted(async () => {
  clusttaVersion.value = await utils.getRawClusttaVersion();
});


</script>

<style scoped>
@import "@/assets/desktop.css";

.close {
  /* background-color: tomato; */
  opacity: .5;
  border-radius: 60px;
  position: absolute;
  top: 8px;
  right: 8px;
}

.close:hover {
  opacity: 1;
  background-color: crimson;
  transform: scale(1.02);

}

.login-button-icon {
  filter: invert(100%);
  width: 20px;
  height: 20px;
}

.login-button-text {
  font-size: 14px;
}

.pop-up-actions {
  /* background-color: red; */
  justify-content: flex-end;
  gap: .5rem;
}

.button {
  width: max-content;
  min-width: 50px;
  padding: .2rem .5rem;
  gap: .2rem;
  height: 40px;
}

.version-info {
  gap: .2rem;
  font-size: 14px;
  width: 100%;
  display: flex;
  flex-direction: column;
  padding: .2rem;
  color: var(--white);
}

.general-container {
  gap: 0px;
  gap: .3rem;
  padding-top: .5rem;
  align-items: flex-start;
  /* background-color: red; */
}

.input-short {
  flex: 1;
  width: 100%;
}

.listbox-short {

  flex: 1;
  width: 130px;
}

.input-label {

  font-family: Inter, sans-serif;
  color: white;
  font-size: 16px;
  white-space: nowrap;
  flex: 1;

}
</style>



