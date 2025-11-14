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
    'davinci resolve': 'drp',
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
    // 3D/Modeling
    '3d modeling': 'cube',
    'sculpting': 'palette',
    'rigging': 'bone',
    'uv mapping': 'texture',
    'retopology': 'shapes',
    
    // Animation
    'character animation': 'man-running',
    'motion graphics': 'video-camera',
    'keyframe animation': 'film-strip',
    'procedural animation': 'flow-chart',
    
    // Texturing/Materials
    'texturing': 'palette',
    'material creation': 'palette',
    'pbr texturing': 'texture',
    'procedural texturing': 'mystery-ball',
    
    // Lighting/Rendering
    'lighting': 'bulb',
    'rendering': 'lamp',
    'shading': 'lamp',
    'compositing': 'four-squares',
    
    // Effects
    'vfx': 'fire',
    'particle effects': 'fire',
    'simulation': 'flow-chart',
    'dynamics': 'shapes',
    
    // 2D/Concept
    'concept art': 'palette',
    'illustration': 'palette',
    'digital painting': 'palette',
    'photo editing': 'image',
    'graphic design': 'shapes',
    
    // Video/Film
    'video editing': 'film-reel',
    'color grading': 'film-strip',
    'cinematography': 'camera',
    'storyboarding': 'open-book',
    
    // Audio
    'sound design': 'drum',
    'music composition': 'music',
    'audio editing': 'music',
    
    // Game Development
    'game design': 'mystery-ball',
    'level design': 'compass',
    'game programming': 'flow-chart',
    
    // Project Management
    'project management': 'boxes',
    'pipeline development': 'flow-chart',
    'technical direction': 'compass',
    'art direction': 'palette'
  };
  
  return skillToIcon[skillNameLower] || 'generic';
};
