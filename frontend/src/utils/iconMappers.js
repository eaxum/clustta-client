/**
 * Icon mapping utilities for tools and skills
 * Maps software names to file icons and skill names to thematic icons
 */

/**
 * Get the file icon path for a tool/software name
 * @param {string} toolName - The name of the tool/software
 * @returns {string} Path to the icon file
 */
export const getToolLogo = (toolName) => {
  const toolNameLower = (toolName || '').toLowerCase();
  
  // Map common software names to file extensions
  const toolToExtension = {
    'blender': 'blend',
    'maya': 'ma',
    'after effects': 'aep',
    'photoshop': 'psd',
    'illustrator': 'ai',
    'premiere pro': 'prproj',
    'substance painter': 'spp',
    'zbrush': 'zpr',
    'cinema 4d': 'c4d',
    '3ds max': 'max',
    'character creator': 'ccproject',
    'iclone': 'iproject',
    'substance designer': 'sbs',
    'krita': 'kra',
    'clip studio paint': 'clip',
    'corel draw': 'cdr',
    'affinity designer': 'afdesign',
    'affinity photo': 'afphoto',
    'affinity publisher': 'afpub',
    'fl studio': 'flp',
    'davinci resolve': 'default',
    'character animator': 'chproj',
    'animate': 'fla',
    'indesign': 'indd',
    'lightroom': 'lrcat',
    'houdini': 'default',
    'unreal engine': 'default',
    'unity': 'default',
    'marvelous designer': 'default',
    'nuke': 'default'
  };
  
  const extension = toolToExtension[toolNameLower];
  return extension ? `/file-icons/${extension}.svg` : '/file-icons/default.svg';
};

/**
 * Get the icon name for a skill
 * @param {string} skillName - The name of the skill
 * @param {string} category - The category of the skill
 * @returns {string} Icon name to be used with iconStore
 */
export const getSkillIcon = (skillName, category) => {
  const skillNameLower = (skillName || '').toLowerCase();
  
  // Map skills to appropriate icons
  const skillToIcon = {
    // Technical
    '3d modeling': 'cube',
    'rigging': 'bone',
    'rendering': 'lamp',
    'vfx': 'fire',
    'compositing': 'flow-chart',
    'game development': 'gamepad',
    'scripting': 'console',
    'technical art': 'shapes',
    'pipeline development': 'flow-chart',
    
    // Artistic
    'animation': 'man-running',
    'texturing': 'texture',
    'lighting': 'bulb',
    'motion graphics': 'video-camera',
    'character design': 'masks',
    'environment design': 'stall',
    'concept art': 'palette',
    'ui/ux design': 'bezier',
    
    // Management
    'project management': 'boxes',
    'art direction': 'palette'
  };
  
  return skillToIcon[skillNameLower] || 'generic';
};
