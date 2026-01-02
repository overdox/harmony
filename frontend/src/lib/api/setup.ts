import { api } from './client';

export interface SetupStatus {
	completed: boolean;
	mediaRoot: string;
}

export interface FolderInfo {
	name: string;
	path: string;
	hasAudio: boolean;
	children: number;
}

export interface BrowseFoldersResponse {
	currentPath: string;
	parentPath: string;
	isMediaRoot: boolean;
	hasAudioFiles: boolean;
	folders: FolderInfo[];
}

export interface SelectedFoldersResponse {
	paths: string[];
}

/**
 * Get setup status
 */
export async function getSetupStatus(): Promise<SetupStatus> {
	const response = await api.get<SetupStatus>('/setup/status');
	return response.data;
}

/**
 * Browse folders in the media root
 */
export async function browseFolders(path?: string): Promise<BrowseFoldersResponse> {
	const response = await api.get<BrowseFoldersResponse>('/setup/folders', { path });
	return response.data;
}

/**
 * Get currently selected folders
 */
export async function getSelectedFolders(): Promise<string[]> {
	const response = await api.get<SelectedFoldersResponse>('/setup/selected-folders');
	return response.data.paths;
}

/**
 * Set selected folders
 */
export async function setSelectedFolders(paths: string[]): Promise<void> {
	await api.post('/setup/selected-folders', { paths });
}

/**
 * Complete setup
 */
export async function completeSetup(startScan: boolean = true): Promise<void> {
	await api.post('/setup/complete', { startScan });
}
