// Slash-command shortcuts for the agent console. Each command maps a short typed phrase
// to a fully-formed natural-language prompt that the LLM treats as if the user had typed it.
// Shortcuts are pure client-side prompt rewriting â€” no new agent tools.

import { ignoreTemplates } from '@/lib/ignoreTemplates';

// help text shown by /help, also used by the autocomplete menu.
export const agentShortcuts = [
  { command: '/setup', args: '(attach a script)', description: 'Scaffold an animation project from the attached script.' },
  { command: '/production', args: '(attach a file)', description: 'Set up an animation production with the attached file as reference.' },
  { command: '/add', args: '<email>', description: 'Add a collaborator to this project.' },
  { command: '/mine', args: '', description: 'Filter the browser to everything assigned to me.' },
  { command: '/assign', args: '@user [filter]', description: 'Bulk-assign matching assets to a user.' },
  { command: '/status', args: '<name> [filter]', description: 'Bulk-change status of matching assets.' },
  { command: '/clear', args: '', description: 'Clear all browser filters.' },
  { command: '/summary', args: '', description: 'Show a project summary (counts by status, type, assignee).' },
  { command: '/export', args: '[all tasks|all blender tasks|all assets|all assets here]', description: 'Preview and export project asset data.' },
  { command: '/ignore', args: '<dcc>', description: 'Apply an ignore-list preset (e.g. Maya, Unreal, Blender).' },
  { command: '/who', args: '', description: 'List everyone on this project as clickable user chips.' },
  { command: '/clustta', args: '', description: 'Show what the Clustta agent can do.' },
  { command: '/help', args: '', description: 'List available shortcuts.' },
];

// Static description of the agent's capabilities, rendered locally by /clustta.
const clusttaCapabilities = [
  'Clustta agent capabilities:',
  '',
  'Project scaffolding:',
  '- Scaffold full animation productions from a script (episodes, sequences, shots, library assets, tasks).',
  '',
  'Browser & filtering:',
  '- Filter assets by assignee, status, type, tag, or path.',
  '- Clear filters and switch between flat and tree views.',
  '',
  'Bulk actions:',
  '- Bulk-assign assets to users.',
  '- Bulk-change status across matching assets.',
  '- Bulk-tag and bulk-rename.',
  '',
  'Project structure:',
  '- Create collections, assets, and tasks.',
  '- Manage asset types, statuses, and tags.',
  '',
  'Collaboration:',
  '- Invite collaborators by email and manage their roles.',
  '- Apply DCC ignore-list presets (Maya, Unreal, Blender, Unity, Houdini, etc.).',
  '',
  'Reporting:',
  '- Project summaries with counts by status, type, and assignee.',
  '- Preview and export filtered or project-wide asset data as JSON, CSV, or plain text.',
  '',
  'Type /help to see available slash-command shortcuts.',
].join('\n');

// Returns an alphabetically-sorted array of preset names from ignoreTemplates.
export function listIgnorePresets() {
  return Object.keys(ignoreTemplates).sort();
}

// Picks the closest ignore preset by case-insensitive substring match. Returns the canonical name or null.
function matchIgnorePreset(query) {
  const q = query.trim().toLowerCase();
  if (!q) return null;
  const presets = listIgnorePresets();
  const exact = presets.find(p => p.toLowerCase() === q);
  if (exact) return exact;
  return presets.find(p => p.toLowerCase().includes(q)) || null;
}

// Renders the /help reply directly without going through the LLM.
function renderHelp() {
  const lines = agentShortcuts.map(s => `- ${s.command} ${s.args} - ${s.description}`);
  return ['Available shortcuts:', ...lines].join('\n');
}

export function isAgentShortcut(command) {
  const normalized = command.toLowerCase();
  return agentShortcuts.some((shortcut) => shortcut.command === normalized)
    || ['/produce', '/invite'].includes(normalized);
}

const restorableShortcutCommands = ['/setup', '/production', '/mine', '/clear', '/who', '/summary'];

export function restoreShortcutDisplay(prompt) {
  for (const command of restorableShortcutCommands) {
    if (expandShortcut(command)?.prompt === prompt) return command;
  }
  return '';
}

// expandShortcut takes raw input and returns either a rewritten prompt string, a synthetic local
// reply (no LLM call), or null when the input is not a shortcut and should be sent verbatim.
export function expandShortcut(rawInput) {
  const text = rawInput.trim();
  if (!text.startsWith('/')) return null;

  const spaceIdx = text.indexOf(' ');
  const command = (spaceIdx === -1 ? text : text.slice(0, spaceIdx)).toLowerCase();
  const rest = spaceIdx === -1 ? '' : text.slice(spaceIdx + 1).trim();

  switch (command) {
    case '/help':
      return { localReply: renderHelp() };

    case '/clustta':
      return { localReply: clusttaCapabilities };

    case '/setup':
      return {
        prompt: rest
          ? `Set up this animation project from the attached script. Additional context: ${rest}`
          : 'Set up this animation project from the attached script.',
      };

    case '/production':
    case '/produce':
      return {
        prompt: rest
          ? `Set up an animation production with the attached file as reference. Additional context: ${rest}`
          : 'Set up an animation production with the attached file as reference.',
      };

    case '/add':
    case '/invite': {
      if (!rest) return { error: 'Usage: /add <email>' };
      return { prompt: `Add ${rest} as a collaborator on this project.` };
    }

    case '/mine':
      return { prompt: 'Filter the browser to show everything assigned to me across the whole project.' };

    case '/assign': {
      if (!rest) return { error: 'Usage: /assign @user [filter]' };
      return { prompt: `Bulk-assign ${rest}.` };
    }

    case '/status': {
      if (!rest) return { error: 'Usage: /status <status-name> [filter]' };
      const parts = rest.split(/\s+/);
      const statusName = parts.shift();
      const filter = parts.join(' ').trim();
      const tail = filter ? ` matching ${filter}` : '';
      return { prompt: `Set status to "${statusName}" for all assets${tail}.` };
    }

    case '/clear':
      return { prompt: 'Clear all browser filters.' };

    case '/who':
      return {
        prompt: 'Call list_users for this project, then reply with each member as a user chip on its own line using the [[user:<id>|Display Name]] token format. Do not include any other text, IDs, or commentary.',
      };

    case '/summary':
      return { prompt: 'Give me a project summary with breakdowns by status, asset type, and assignee.' };

    case '/export': {
      const scopes = {
        '': 'selection',
        'all tasks': 'all_tasks',
        'all blender tasks': 'blender_tasks',
        'all assets': 'all_assets',
        'all assets here': 'selection',
      };
      const scope = scopes[rest.toLowerCase()];
      if (!scope) {
        return { error: 'Usage: /export [all tasks|all blender tasks|all assets|all assets here]' };
      }
      return { action: 'export', scope };
    }

    case '/ignore': {
      if (!rest) {
        return { error: `Usage: /ignore <dcc>. Available presets: ${listIgnorePresets().join(', ')}` };
      }
      const preset = matchIgnorePreset(rest);
      if (!preset) {
        return { error: `Unknown ignore preset "${rest}". Available presets: ${listIgnorePresets().join(', ')}` };
      }
      const patterns = ignoreTemplates[preset];
      const list = patterns.map(p => `"${p}"`).join(', ');
      return {
        prompt: `Add the following ignore patterns to this project for the ${preset} preset: ${list}. Use the add_ignore_pattern tool for each.`,
      };
    }

    default:
      return { error: `Unknown shortcut "${command}". Type /help to list available shortcuts.` };
  }
}
