import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpContext } from '@angular/common/http';

import { environment } from '../../environments/environment';
import { BYPASS_GLOBAL_ERROR } from '../interceptors/error.interceptor';

export interface MusicList {
  id: string;
  name: string;
  album_count: number;
  created_at: string;
  updated_at: string;
}

export interface ListItem {
  album_id: string;

  // backend may return one of these
  album_title?: string;
  title?: string;

  artist_name?: string;
  album_cover?: string;

  added_at: string;
}

@Injectable({
  providedIn: 'root',
})
export class CustomListService {
  private http = inject(HttpClient);
  private apiUrl = environment.apiUrl;

  getLists() {
    return this.http.get<MusicList[]>(
      `${this.apiUrl}/api/lists`,
      {
        withCredentials: true,
      },
    );
  }

  createList(name: string) {
    return this.http.post(
      `${this.apiUrl}/api/lists`,
      {
        name,
      },
      {
        withCredentials: true,
      },
    );
  }

  deleteList(id: string) {
    return this.http.delete(
      `${this.apiUrl}/api/lists/${id}`,
      {
        withCredentials: true,
      },
    );
  }

  getListItems(id: string) {
    return this.http.get<ListItem[]>(
      `${this.apiUrl}/api/lists/${id}/items`,
      {
        withCredentials: true,
      },
    );
  }

  addAlbumToList(listId: string, albumId: string) {
    return this.http.post(
      `${this.apiUrl}/api/lists/${listId}/items`,
      {
        album_id: albumId,
      },
      {
        withCredentials: true,
        context: new HttpContext().set(BYPASS_GLOBAL_ERROR, true),
      },
    );
  }

  removeAlbumFromList(listId: string, albumId: string) {
    return this.http.delete(
      `${this.apiUrl}/api/lists/${listId}/items/${albumId}`,
      {
        withCredentials: true,
      },
    );
  }
}