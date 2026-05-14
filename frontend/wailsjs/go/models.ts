export namespace main {
  export class SpotifyMetadataRequest {
    url = "";
    batch = false;
    delay = 0;
    timeout = 0;
    separator = "";

    constructor(source: Partial<SpotifyMetadataRequest> = {}) {
      Object.assign(this, source);
    }
  }

  export class SpotifySearchRequest {
    query = "";
    limit = 0;

    constructor(source: Partial<SpotifySearchRequest> = {}) {
      Object.assign(this, source);
    }
  }

  export class SpotifySearchByTypeRequest {
    query = "";
    search_type = "";
    limit = 0;
    offset = 0;

    constructor(source: Partial<SpotifySearchByTypeRequest> = {}) {
      Object.assign(this, source);
    }
  }

  export class DownloadRequest {
    constructor(source: Record<string, unknown> = {}) {
      Object.assign(this, source);
    }
  }
}

export namespace backend {
  export class DownloadQueueInfo {
    queue: DownloadQueueItem[] = [];
    active = 0;
    pending = 0;
    completed = 0;
    failed = 0;
    skipped = 0;
    total = 0;

    constructor(source: Partial<DownloadQueueInfo> = {}) {
      Object.assign(this, source);
    }
  }

  export interface DownloadQueueItem {
    id: string;
    track_name: string;
    artist_name: string;
    album_name: string;
    spotify_id: string;
    status: string;
    error_message?: string;
    file_path?: string;
    size_mb?: number;
  }

  export class SearchResponse {
    results: SearchResult[] = [];
    total = 0;

    constructor(source: Partial<SearchResponse> = {}) {
      Object.assign(this, source);
    }
  }

  export interface SearchResult {
    id: string;
    name: string;
    artists?: string;
    external_urls: string;
    images?: string;
    type?: string;
  }

  export interface FileInfo {
    name: string;
    path: string;
    is_dir: boolean;
    size: number;
    modified: string;
  }

  export interface RenamePreview {
    old_path: string;
    new_path: string;
    success: boolean;
    error?: string;
  }

  export interface RenameResult {
    old_path: string;
    new_path: string;
    success: boolean;
    error?: string;
  }

  export interface AudioMetadata {
    title?: string;
    artist?: string;
    album?: string;
    album_artist?: string;
    track_number?: string;
    year?: string;
  }
}
