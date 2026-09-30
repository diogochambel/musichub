import { Injectable, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

export interface Notification {
  id: string;
  user_id: string;
  request_id: string;
  album_title: string;
  response: 'aceite' | 'recusado';
  read: boolean;
  created_at: string;
}

@Injectable({
  providedIn: 'root',
})
export class NotificationService {
  private http = inject(HttpClient);
  private apiUrl = environment.apiUrl;

  unreadCount = signal(0);

  getNotifications() {
    return this.http.get<Notification[]>(`${this.apiUrl}/api/notifications`, {
      withCredentials: true,
    });
  }

  getUnreadCount() {
    return this.http.get<{ unread: number }>(`${this.apiUrl}/api/notifications/unread-count`, {
      withCredentials: true,
    });
  }

  refreshUnreadCount() {
    this.getUnreadCount().subscribe({
      next: (res) => this.unreadCount.set(res.unread),
      error: () => this.unreadCount.set(0),
    });
  }

  markAllRead() {
    return this.http.post(
      `${this.apiUrl}/api/notifications/mark-all-read`,
      {},
      { withCredentials: true },
    );
  }

  markAsRead(id: string) {
    return this.http.post(
      `${this.apiUrl}/api/notifications/${id}/read`,
      {},
      { withCredentials: true },
    );
  }
}
