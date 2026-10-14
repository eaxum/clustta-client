const ASSET_ENTITY = 'asset';
const SHOT_ENTITY = 'shot';

const ASSET_PLACEHOLDER = '<Asset>';
const SHOT_PLACEHOLDER = '<Shot>';

export const inferDirectoryTemplateEntity = (id, template = '') => {
  const value = `${id || ''} ${template || ''}`.toLowerCase();
  if (value.includes(SHOT_PLACEHOLDER.toLowerCase()) || value.includes(SHOT_ENTITY)) return SHOT_ENTITY;
  if (value.includes(ASSET_PLACEHOLDER.toLowerCase()) || value.includes(ASSET_ENTITY)) return ASSET_ENTITY;
  return '';
};

export const getDefaultTaskOutput = (template, taskType) => {
  const entity = inferDirectoryTemplateEntity(template.id, template.template);
  const subjectPlaceholder = entity === SHOT_ENTITY ? SHOT_PLACEHOLDER : ASSET_PLACEHOLDER;
  return `${subjectPlaceholder}-${String(taskType || '').toLowerCase()}`;
};

export const getDirectoryTemplatePreviewValues = (template) => {
  const entity = inferDirectoryTemplateEntity(template.id, template.template);
  const isShotTemplate = entity === SHOT_ENTITY;
  const taskType = isShotTemplate ? 'animation' : 'modeling';
  const subject = isShotTemplate ? 'shot010' : 'hero';

  return {
    Episode: 'ep01',
    Sequence: 'seq001',
    Shot: 'shot010',
    CollectionType: 'character',
    Category: 'heroes',
    Asset: 'hero',
    AssetType: taskType,
    OutputName: `${subject}-${taskType}`,
    TemplateExtension: '.ext',
  };
};
