import { Injectable, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { environment } from '../../environments/environment';
import { LoginResponse, RegisterResponse, TokenData } from '../models/user.interface';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private http: HttpClient;
  private router: Router;
  private apiUrl = environment.apiUrl;

  private loggedIn = signal(this.hasToken());
  isLoggedIn = computed(() => this.loggedIn());
  username = signal(this.getStoredUsername());

  constructor(http: HttpClient, router: Router) {
    this.http = http;
    this.router = router;
  }

  login(username: string, password: string) {
    return this.http.post<LoginResponse>(`${this.apiUrl}/api/auth/login`, {
      username,
      password,
    });
  }

  register(username: string, email: string, password: string, birth_date: string) {
    return this.http.post<RegisterResponse>(`${this.apiUrl}/api/auth/register`, {
      username,
      email,
      password,
      birth_date,
    });
  }

  handleLoginSuccess(token: string): void {
    localStorage.setItem('token', token);

    const tokenData = this.decodeToken(token);
    if (tokenData) {
      localStorage.setItem('user_id', tokenData.user_id);
      localStorage.setItem('username', tokenData.username);
      this.username.set(tokenData.username);
    }

    this.loggedIn.set(true);
  }

  logout(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('user_id');
    localStorage.removeItem('username');

    this.loggedIn.set(false);
    this.username.set('');
    this.router.navigate(['/']);
  }

  getToken(): string | null {
    return localStorage.getItem('token');
  }

  getCurrentUserId(): string | null {
    return localStorage.getItem('user_id');
  }

  getCurrentUsername(): string | null {
    return localStorage.getItem('username');
  }

  private hasToken(): boolean {
    return !!localStorage.getItem('token');
  }

  private getStoredUsername(): string {
    return localStorage.getItem('username') || '';
  }

  private decodeToken(token: string): TokenData | null {
    try {
      const payload = token.split('.')[0];
      const decoded = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
      return JSON.parse(decoded) as TokenData;
    } catch {
      return null;
    }
  }
}
