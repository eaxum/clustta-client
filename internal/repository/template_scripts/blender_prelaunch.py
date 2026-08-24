from pathlib import Path

import bpy


scene = bpy.context.scene
scene_name = Path(CLUSTTA_CONTEXT["target_file"]).stem
scene.name = scene_name

collection = bpy.data.collections.get(scene_name)
if collection is None:
    collection = bpy.data.collections.new(scene_name)
if collection.name not in {item.name for item in scene.collection.children}:
    scene.collection.children.link(collection)

camera_name = f"{scene_name}-action"
camera = bpy.data.objects.get(camera_name)
if camera is None:
    camera_data = bpy.data.cameras.new(camera_name)
    camera = bpy.data.objects.new(camera_name, camera_data)
    collection.objects.link(camera)

scene.camera = camera
scene.frame_start = 101
scene.frame_end = 351
