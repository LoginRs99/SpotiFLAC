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
    [key: string]: any;

    constructor(source: any = {}) {
      Object.assign(this, source);
    }
  }

  export class LyricsDownloadRequest {
    [key: string]: any;

    constructor(source: any = {}) {
      Object.assign(this, source);
    }
  }

  export class CoverDownloadRequest {
    [key: string]: any;

    constructor(source: any = {}) {
      Object.assign(this, source);
    }
  }

  export class HeaderDownloadRequest {
    [key: string]: any;

    constructor(source: any = {}) {
      Object.assign(this, source);
    }
  }

  export class GalleryImageDownloadRequest {
    [key: string]: any;

    constructor(source: any = {}) {
      Object.assign(this, source);
    }
  }

  export class AvatarDownloadRequest {
    [key: string]: any;

    constructor(source: any = {}) {
      Object.assign(this, source);
    }
  }
}

export namespace backend {
  export class DownloadQueueInfo {
    [key: string]: any;

    queue: DownloadQueueItem[] = [];
    is_downloading = false;
    queued_count = 0;
    completed_count = 0;
    failed_count = 0;
    skipped_count = 0;
    total_downloaded = 0;
    current_speed = 0;
    session_start_time = 0;
    total = 0;

    constructor(source: Partial<DownloadQueueInfo> = {}) {
      Object.assign(this, source);
    }
  }

  export interface DownloadQueueItem {
    [key: string]: any;

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
    [key: string]: any;

    tracks: SearchResult[] = [];
    albums: SearchResult[] = [];
    artists: SearchResult[] = [];
    playlists: SearchResult[] = [];
    results: SearchResult[] = [];
    total = 0;

    constructor(source: Partial<SearchResponse> = {}) {
      Object.assign(this, source);
    }
  }

  export interface SearchResult {
    [key: string]: any;

    id: string;
    name: string;
    artists?: string;
    external_urls: string;
    images?: string;
    type?: string;
  }

  export interface FileInfo {
    [key: string]: any;

    name: string;
    path: string;
    is_dir: boolean;
    size: number;
    modified: string;
  }

  export interface RenamePreview {
    [key: string]: any;

    old_name?: string;
    new_name?: string;
    old_path: string;
    new_path: string;
    success: boolean;
    error?: string;
  }

  export interface RenameResult {
    [key: string]: any;

    old_path: string;
    new_path: string;
    success: boolean;
    error?: string;
  }

  export interface AudioMetadata {
    [key: string]: any;

    title?: string;
    artist?: string;
    album?: string;
    album_artist?: string;
    track_number?: number;
    disc_number?: number;
    year?: string;
  }
}
