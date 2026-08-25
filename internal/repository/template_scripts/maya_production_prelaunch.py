import os

import maya.cmds as cmds


WORKSPACE_FILE_NAME = "workspace.mel"


def normalize_path(path):
    return os.path.normcase(os.path.realpath(os.path.normpath(path)))


def is_within_project(target_file, project_root):
    try:
        return os.path.commonpath([project_root, target_file]) == project_root
    except ValueError:
        return False


project_root = CLUSTTA_CONTEXT["project_root"]
target_file = CLUSTTA_CONTEXT["target_file"]
workspace_file = os.path.join(project_root, WORKSPACE_FILE_NAME)

if not os.path.isfile(workspace_file):
    raise RuntimeError(f"Maya workspace file not found: {workspace_file}")

expected_workspace = normalize_path(project_root)
if not is_within_project(normalize_path(target_file), expected_workspace):
    raise RuntimeError("Maya file is outside the Clustta project workspace")

cmds.workspace(project_root, openWorkspace=True)

active_workspace = normalize_path(cmds.workspace(query=True, active=True))
if active_workspace != expected_workspace:
    raise RuntimeError("Maya workspace validation failed")
