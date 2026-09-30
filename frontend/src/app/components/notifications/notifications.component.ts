import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { finalize } from 'rxjs';

import { NotificationService, Notification } from '../../services/notification.service';

@Component({
  selector: 'app-notifications',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './notifications.component.html',
  styleUrls: ['./notifications.component.css'],
})
export class NotificationsComponent implements OnInit {
  private notificationService: NotificationService = inject(NotificationService);

  notifications = signal<Notification[]>([]);
  loading = signal(true);

  ngOnInit(): void {
    this.loadNotifications();
  }

  loadNotifications(): void {
    this.loading.set(true);

    this.notificationService
      .getNotifications()
      .pipe(finalize(() => this.loading.set(false)))
      .subscribe({
        next: (notifications: Notification[]) => {
          this.notifications.set(notifications ?? []);
        },
        error: (err) => {
          console.error('Failed to load notifications', err);
          this.notifications.set([]);
        },
      });
  }

  markAllRead(): void {
    this.notificationService.markAllRead().subscribe({
      next: () => {
        this.notifications.update((items) =>
          items.map((notification) => ({
            ...notification,
            read: true,
          })),
        );
        this.notificationService.refreshUnreadCount();
      },
    });
  }

  responseLabel(response: string): string {
    return response === 'aceite' ? 'Accepted' : 'Rejected';
  }

  markAsRead(notification: Notification): void {
    if (notification.read) return;

    this.notificationService.markAsRead(notification.id).subscribe(() => {
      notification.read = true;
      this.notificationService.refreshUnreadCount();
    });
  }
}
