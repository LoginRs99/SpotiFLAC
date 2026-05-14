async function apiJSON<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(path, {
    headers: { "Content-Type": "application/json", ...(init.headers ?? {}) },
    ...init,
  });
  if (!response.ok) {
    let message = response.statusText;
    try {
      const payload = await response.json();
      message = payload.error || message;
    } catch {
      message = await response.text();
    }
    throw new Error(message);
  }
  return response.json() as Promise<T>;
}

function postJSON<T>(path: string, body?: unknown): Promise<T> {
  return apiJSON<T>(path, { method: "POST", body: JSON.stringify(body ?? {}) });
}

function deleteJSON<T>(path: string): Promise<T> {
  return apiJSON<T>(path, { method: "DELETE" });
}

function filenameFromDisposition(header: string | null): string {
  if (!header) return "SpotiFLAC-download";
  const utf8 = header.match(/filename\*=UTF-8''([^;]+)/i);
  if (utf8?.[1]) return decodeURIComponent(utf8[1]);
  const plain = header.match(/filename="?([^";]+)"?/i);
  return plain?.[1] ?? "SpotiFLAC-download";
}

function saveBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  window.setTimeout(() => URL.revokeObjectURL(url), 5000);
}

export function GetSpotifyMetadata(req: unknown): Promise<string> {
  return postJSON<unknown>("/api/metadata", req).then(JSON.stringify);
}

export function GetCurrentIPInfo(): Promise<string> {
  return apiJSON<unknown>("/api/current-ip").then(JSON.stringify);
}

export async function DownloadTrack(req: unknown): Promise<Record<string, unknown>> {
  const response = await fetch("/api/download", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req ?? {}),
  });
  if (!response.ok) {
    let message = response.statusText;
    try {
      message = (await response.json()).error || message;
    } catch {
      message = await response.text();
    }
    throw new Error(message);
  }

  const blob = await response.blob();
  const filename = filenameFromDisposition(response.headers.get("Content-Disposition"));
  saveBlob(blob, filename);

  const metadataHeader = response.headers.get("X-SpotiFLAC-Response");
  if (metadataHeader) {
    return JSON.parse(metadataHeader) as Record<string, unknown>;
  }
  return { success: true, message: "Download completed successfully", file: filename };
}

export function SearchSpotify(req: unknown): Promise<unknown> {
  return postJSON("/api/search", req);
}

export function SearchSpotifyByType(req: unknown): Promise<unknown[]> {
  return postJSON("/api/search-by-type", req);
}

export function GetStreamingURLs(spotifyTrackID: string, region: string): Promise<string> {
  const params = new URLSearchParams({ spotifyTrackID, region });
  return apiJSON<unknown>(`/api/streaming-urls?${params}`).then(JSON.stringify);
}

export function GetDefaults(): Promise<Record<string, string>> {
  return apiJSON("/api/defaults");
}

export function LoadSettings(): Promise<Record<string, unknown>> {
  return apiJSON("/api/settings");
}

export function SaveSettings(settings: Record<string, unknown>): Promise<void> {
  return postJSON<void>("/api/settings", settings);
}

export function LoadFonts(): Promise<Record<string, unknown>[]> {
  return apiJSON("/api/fonts");
}

export function SaveFonts(fonts: Record<string, unknown>[]): Promise<void> {
  return postJSON<void>("/api/fonts", fonts);
}

export function GetDownloadProgress(): Promise<unknown> {
  return apiJSON("/api/download/progress");
}

export function GetDownloadQueue(): Promise<unknown> {
  return apiJSON("/api/download/queue");
}

export function AddToDownloadQueue(spotifyID: string, trackName: string, artistName: string, albumName: string): Promise<string> {
  return postJSON<{ itemID: string }>("/api/download/queue/add", { spotifyID, trackName, artistName, albumName }).then((r) => r.itemID);
}

export function MarkDownloadItemFailed(itemID: string, errorMsg: string): Promise<void> {
  return postJSON<void>("/api/download/queue/failed", { itemID, errorMsg });
}

export function CancelAllQueuedItems(): Promise<void> {
  return postJSON<void>("/api/download/queue/cancel");
}

export function ClearCompletedDownloads(): Promise<void> {
  return postJSON<void>("/api/download/queue/clear-completed");
}

export function ClearAllDownloads(): Promise<void> {
  return postJSON<void>("/api/download/queue/clear-all");
}

export function SkipDownloadItem(itemID: string, filePath: string): Promise<void> {
  return postJSON<void>("/api/download/queue/skip", { itemID, filePath });
}

export function CheckFilesExistence(outputDir: string, rootDir: string, tracks: unknown[]): Promise<unknown[]> {
  return postJSON("/api/download/check-files", { outputDir, rootDir, tracks });
}

export function CreateM3U8File(playlistName: string, outputDir: string, filePaths: string[]): Promise<void> {
  return postJSON<void>("/api/download/m3u8", { playlistName, outputDir, filePaths });
}

export function GetDownloadHistory(): Promise<unknown[]> {
  return apiJSON("/api/history/downloads");
}

export function ClearDownloadHistory(): Promise<void> {
  return deleteJSON<void>("/api/history/downloads");
}

export function DeleteDownloadHistoryItem(id: string): Promise<void> {
  return deleteJSON<void>(`/api/history/downloads/item?id=${encodeURIComponent(id)}`);
}

export function GetFetchHistory(): Promise<unknown[]> {
  return apiJSON("/api/history/fetches");
}

export function AddFetchHistory(item: unknown): Promise<void> {
  return postJSON<void>("/api/history/fetches", item);
}

export function ClearFetchHistory(): Promise<void> {
  return deleteJSON<void>("/api/history/fetches");
}

export function DeleteFetchHistoryItem(id: string): Promise<void> {
  return deleteJSON<void>(`/api/history/fetches/item?id=${encodeURIComponent(id)}`);
}

export function ClearFetchHistoryByType(type: string): Promise<void> {
  return deleteJSON<void>(`/api/history/fetches/type?type=${encodeURIComponent(type)}`);
}

export function GetRecentFetches(): Promise<string> {
  return apiJSON<string>("/api/history/recent-fetches");
}

export function SaveRecentFetches(payload: string): Promise<void> {
  return postJSON<void>("/api/history/recent-fetches", { payload });
}

export function CheckAPIStatus(apiType: string, apiURL: string): Promise<boolean> {
  const params = new URLSearchParams({ type: apiType, url: apiURL });
  return apiJSON<{ online: boolean }>(`/api/status/check?${params}`).then((r) => r.online);
}

export function CheckCustomTidalAPI(apiURL: string): Promise<boolean> {
  return apiJSON<{ online: boolean }>(`/api/status/custom-tidal?url=${encodeURIComponent(apiURL)}`).then((r) => r.online);
}

export function CheckFFmpegInstalled(): Promise<boolean> {
  return apiJSON<{ installed: boolean }>("/api/status/ffmpeg").then((r) => r.installed);
}

export const IsFFmpegInstalled = CheckFFmpegInstalled;

export function DownloadFFmpeg(): Promise<unknown> {
  return postJSON("/api/status/ffmpeg/download");
}

export function CheckTrackAvailability(spotifyTrackID: string): Promise<string> {
  return apiJSON<unknown>(`/api/availability?spotifyTrackID=${encodeURIComponent(spotifyTrackID)}`).then(JSON.stringify);
}

export function DownloadLyrics(req: unknown): Promise<unknown> {
  return postJSON("/api/lyrics", req);
}

export function DownloadCover(req: unknown): Promise<unknown> {
  return postJSON("/api/cover", req);
}

export function DownloadHeader(req: unknown): Promise<unknown> {
  return postJSON("/api/header", req);
}

export function DownloadGalleryImage(req: unknown): Promise<unknown> {
  return postJSON("/api/gallery-image", req);
}

export function DownloadAvatar(req: unknown): Promise<unknown> {
  return postJSON("/api/avatar", req);
}

export function GetPreviewURL(trackID: string): Promise<string> {
  return apiJSON<{ url: string }>(`/api/preview?trackID=${encodeURIComponent(trackID)}`).then((r) => r.url);
}

export function GetTrackISRC(spotifyTrackID: string): Promise<string> {
  return apiJSON<{ isrc: string }>(`/api/isrc?spotifyTrackID=${encodeURIComponent(spotifyTrackID)}`).then((r) => r.isrc);
}

function unsupported<T>(name: string, fallback: T): Promise<T> {
  console.warn(`${name} is not available in the Docker web build`);
  return Promise.resolve(fallback);
}

export function OpenFolder(_path: string): Promise<void> {
  return unsupported("OpenFolder", undefined);
}

export function OpenConfigFolder(): Promise<void> {
  return unsupported("OpenConfigFolder", undefined);
}

export function SelectFolder(defaultPath = ""): Promise<string> {
  return unsupported("SelectFolder", defaultPath);
}

export function SelectFile(): Promise<string> {
  return unsupported("SelectFile", "");
}

export function SelectAudioFiles(): Promise<string[]> {
  return unsupported("SelectAudioFiles", []);
}

export function ListAudioFilesInDir(_dirPath: string): Promise<unknown[]> {
  return unsupported("ListAudioFilesInDir", []);
}

export function ListDirectoryFiles(_dirPath: string): Promise<unknown[]> {
  return unsupported("ListDirectoryFiles", []);
}

export function ReadFileMetadata(_filePath: string): Promise<unknown> {
  return unsupported("ReadFileMetadata", {});
}

export function PreviewRenameFiles(_files: string[], _format: string): Promise<unknown[]> {
  return unsupported("PreviewRenameFiles", []);
}

export function RenameFilesByMetadata(_files: string[], _format: string): Promise<unknown[]> {
  return unsupported("RenameFilesByMetadata", []);
}

export function ConvertAudio(_req: unknown): Promise<unknown[]> {
  return unsupported("ConvertAudio", []);
}

export function ResampleAudio(_req: unknown): Promise<unknown[]> {
  return unsupported("ResampleAudio", []);
}

export function GetFileSizes(_files: string[]): Promise<Record<string, number>> {
  return unsupported("GetFileSizes", {});
}

export function SaveSpectrumImage(_audioFilePath: string, _base64Data: string): Promise<string> {
  return unsupported("SaveSpectrumImage", "");
}

export function GetFlacInfoBatch(_paths: string[]): Promise<unknown[]> {
  return unsupported("GetFlacInfoBatch", []);
}

export function ReadTextFile(_filePath: string): Promise<string> {
  return unsupported("ReadTextFile", "");
}

export function ReadFileAsBase64(_filePath: string): Promise<string> {
  return unsupported("ReadFileAsBase64", "");
}

export function DecodeAudioForAnalysis(_filePath: string): Promise<unknown> {
  return unsupported("DecodeAudioForAnalysis", {});
}

export function ExportFailedDownloads(): Promise<string> {
  return unsupported("ExportFailedDownloads", "Export is not available in the Docker web build.");
}
