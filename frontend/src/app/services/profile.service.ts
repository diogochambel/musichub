import { Injectable, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { User, UpdateUserRequest } from '../models/user.interface';
import { AuthService } from './auth.service';
import { tap } from 'rxjs/operators';

@Injectable({ providedIn: 'root' })
export class ProfileService {
  private http = inject(HttpClient);
  private authService = inject(AuthService);
  private apiUrl = environment.apiUrl;

  private userProfile = signal<User | null>(null);
  profile = this.userProfile.asReadonly();

  getProfile(): User | null {
    return this.userProfile();
  }

  fetchProfile() {
    const userId = this.authService.getCurrentUserId();
    if (!userId) {
      this.userProfile.set(null);
      return;
    }

    return this.http
      .get<User>(`${this.apiUrl}/users/${userId}`, {
        withCredentials: true,
      })
      .pipe(
        tap((user) => {
          this.userProfile.set(user);
        }),
      );
  }

  updateProfile(data: UpdateUserRequest) {
    const userId = this.authService.getCurrentUserId();
    if (!userId) return;

    return this.http
      .put<User>(`${this.apiUrl}/users/${userId}`, data, {
        withCredentials: true,
      })
      .pipe(
        tap((updatedUser) => {
          this.userProfile.set(updatedUser);
        }),
      );
  }

  clearProfile(): void {
    this.userProfile.set(null);
  }
}
