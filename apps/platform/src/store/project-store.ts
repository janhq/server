import Cookies from 'js-cookie';
import { create } from 'zustand';

export interface Project {
  id: string;
  name: string;
  createdAt: string;
}

interface ProjectState {
  projects: Project[];
  currentProject: Project | null;
  addProject: (name: string) => void;
  deleteProject: (id: string) => void;
  setCurrentProject: (project: Project) => void;
}

const COOKIE_NAME = 'project-storage';

// Helper to get initial state from cookie (only runs on client)
const getInitialState = (): Pick<ProjectState, 'projects' | 'currentProject'> => {
  // Skip on server-side
  if (typeof window === 'undefined') {
    return {
      projects: [{ id: '1', name: 'Default project', createdAt: new Date().toISOString() }],
      currentProject: { id: '1', name: 'Default project', createdAt: new Date().toISOString() },
    };
  }

  const cookie = Cookies.get(COOKIE_NAME);
  if (cookie) {
    try {
      const parsed = JSON.parse(cookie);
      return {
        projects: parsed.projects || [
          { id: '1', name: 'Default project', createdAt: new Date().toISOString() },
        ],
        currentProject:
          parsed.currentProject || {
            id: '1',
            name: 'Default project',
            createdAt: new Date().toISOString(),
          },
      };
    } catch {
      return {
        projects: [{ id: '1', name: 'Default project', createdAt: new Date().toISOString() }],
        currentProject: { id: '1', name: 'Default project', createdAt: new Date().toISOString() },
      };
    }
  }
  return {
    projects: [{ id: '1', name: 'Default project', createdAt: new Date().toISOString() }],
    currentProject: { id: '1', name: 'Default project', createdAt: new Date().toISOString() },
  };
};

// Helper to save state to cookie
const saveStateToCookie = (state: Pick<ProjectState, 'projects' | 'currentProject'>) => {
  Cookies.set(COOKIE_NAME, JSON.stringify(state), { expires: 365 });
};

export const useProjectStore = create<ProjectState>()((set, get) => ({
  ...getInitialState(),
  addProject: (name: string) => {
    const newProject: Project = {
      id: Date.now().toString(),
      name,
      createdAt: new Date().toISOString(),
    };
    const newState = {
      projects: [...get().projects, newProject],
      currentProject: get().currentProject,
    };
    set(newState);
    saveStateToCookie(newState);
  },
  deleteProject: (id: string) => {
    const projects = get().projects.filter((p) => p.id !== id);
    const currentProject = get().currentProject;

    // If deleting current project, switch to first available
    const newCurrentProject =
      currentProject?.id === id ? (projects[0] || null) : currentProject;

    const newState = {
      projects,
      currentProject: newCurrentProject,
    };
    set(newState);
    saveStateToCookie(newState);
  },
  setCurrentProject: (project: Project) => {
    const newState = {
      projects: get().projects,
      currentProject: project,
    };
    set(newState);
    saveStateToCookie(newState);
  },
}));
