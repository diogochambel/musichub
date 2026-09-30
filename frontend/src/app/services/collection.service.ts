import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { AuthService } from './auth.service';

export interface CollectionItemResponse {
  id: string;
  album_id: string;
  album_title: string;
  release_year: number;
  artist_name: string;
  ean_13: string;
  support: string;
  version_name?: string;
  added_at: string;
}

export interface AddToCollectionRequest {
  album_id: string;
  ean_13: string;
  support: string;
  version_name?: string;
}

@Injectable({ providedIn: 'root' })
export class CollectionService {
  private http = inject(HttpClient);
  private auth = inject(AuthService);
  private apiUrl = environment.apiUrl;

  private authHeaders(): HttpHeaders {
    return new HttpHeaders({ Authorization: `Bearer ${this.auth.getToken()}` });
  }

  getMyCollection() {
    return this.http.get<CollectionItemResponse[]>(`${this.apiUrl}/api/collections`, {
      headers: this.authHeaders(),
    });
  }

  addToCollection(item: AddToCollectionRequest) {
    return this.http.post<any>(`${this.apiUrl}/api/collections`, item, {
      headers: this.authHeaders(),
    });
  }

  removeFromCollection(id: string) {
    return this.http.delete<any>(`${this.apiUrl}/api/collections/${id}`, {
      headers: this.authHeaders(),
    });
  }
}
