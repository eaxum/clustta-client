import os

import maya.cmds as cmds


project_root = CLUSTTA_CONTEXT["project_root"]
cmds.workspace(project_root, openWorkspace=True)

active_workspace = os.path.normcase(os.path.normpath(cmds.workspace(query=True, active=True)))
expected_workspace = os.path.normcase(os.path.normpath(project_root))
if active_workspace != expected_workspace:
    raise RuntimeError("Maya workspace validation failed")
