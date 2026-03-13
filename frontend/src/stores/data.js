import { defineStore } from 'pinia'

export const useDataStore = defineStore('data', {

    state: () => ({

        resource: {
          "name": "Christmas",
          "id": "bbe4e58d-d2d7-4620-9110-10342bd99f20",

          "resource_type_name": "generic",
          "resource_type_icon": "generic",
          "resource_type_id": "c432bcb8-204a-4a95-878f-8613b7cb9b05",
          "collection_id": "404330b0-a33e-4643-9c7d-31f4f7b498a4",
          
      },
        asset:{
          "name": "Christmas",
          "id": "7550e127-b977-4914-ace9-18337624360d",

          "asset_type_name": "generic",
          "asset_type_icon": "generic",
          "asset_type_id": "b463f33e-67c2-4cec-a96c-825148ada3af",
          "collection_id": "db7c99ba-067c-4590-be98-73389f7bb359",
      },
      workflowTemplate: {
        "resources": [],
        "collections": [
          {
            name: "3D Animation",
            id: "2e4fda8c-e170-447f-b455-e8f6e273d6e4",
            is_root: "true",

            collection_type_name: "sequence",
            collection_type_icon: "sequence",
            collection_type_id: "a68e5502-f278-4cf8-8622-ab7c368d2705",
            parent_id: "",
            type: "collection",
          },
          {
            name: "Renders",
            id: "c432bcb8-204a-4a95-878f-8613b7cb9b05",

            collection_type_name: "folder",
            collection_type_icon: "folder",
            collection_type_id: "bbe4e58d-d2d7-4620-9110-10342bd99f20",
            parent_id: "2e4fda8c-e170-447f-b455-e8f6e273d6e4",
          }
        ],
        "assets": [
            {
              name: "Animation",
              id: "8bfc553c-e7ec-417d-bdfd-a4f0d3952872",
    
              asset_type_name: "animation",
              asset_type_icon: "animation",
              asset_type_id: "b463f33e-67c2-4cec-a96c-825148ada3af",
              collection_id: "2e4fda8c-e170-447f-b455-e8f6e273d6e4",
              icon : '/file-icons/blend.svg'
            },
            {
              name: "Layout",
              id: "9ae6a067-db45-4022-9259-92c4148553a2",
    
              asset_type_name: "rigging",
              asset_type_icon: "rigging",
              asset_type_id: "b463f33e-67c2-4cec-a96c-825148ada3af",
              collection_id: "2e4fda8c-e170-447f-b455-e8f6e273d6e4",
              icon : '/file-icons/afdesign.svg'
            },
            {
              name: "Lighting",
              id: "f44d4a44-3119-410a-8062-ae63e38dedde",
    
              asset_type_name: "lighting",
              asset_type_icon: "lighting",
              asset_type_id: "b463f33e-67c2-4cec-a96c-825148ada3af",
              collection_id: "2e4fda8c-e170-447f-b455-e8f6e273d6e4",
              icon : '/file-icons/blend.svg'
            },
        ]
    }

    })

})
