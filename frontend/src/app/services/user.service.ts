import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { User } from '../models/user.interface';

@Injectable({ providedIn: 'root' })
export class UserService {
  private http = inject(HttpClient);
  private apiUrl = environment.apiUrl;

  getUser(userId: string) {
    return this.http.get<User>(`${this.apiUrl}/users/${userId}`);
  }

  getFavoriteArtistId(userId: string) {
    return this.http.get<User>(`${this.apiUrl}/users/${userId}`);
  }

  setFavoriteArtist(userId: string, artistId: string) {
    return this.http.put(`${this.apiUrl}/users/${userId}/favorite-artist`, {
      artist_id: artistId,
    });
  }

  removeFavoriteArtist(userId: string) {
    return this.http.delete(`${this.apiUrl}/users/${userId}/favorite-artist`);
  }
}
