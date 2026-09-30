import { Component, OnInit, effect, inject, signal } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';

import { AuthService } from '../../services/auth.service';
import { ThemeService } from '../../services/theme.service';
import { NotificationService } from '../../services/notification.service';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [RouterLink, RouterLinkActive],
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.css'],
})
export class HeaderComponent implements OnInit {
  menuOpen = false;

  private authService = inject(AuthService);
  private themeService = inject(ThemeService);
  private notificationService = inject(NotificationService);

  isLoggedIn = this.authService.isLoggedIn;
  username = this.authService.username;
  theme = this.themeService.theme;

  unreadCount = this.notificationService.unreadCount;

  constructor() {
    effect(() => {
      if (this.isLoggedIn()) {
        this.notificationService.refreshUnreadCount();
      } else {
        this.notificationService.unreadCount.set(0);
      }
    });
  }

  ngOnInit(): void {
    if (this.isLoggedIn()) {
      this.notificationService.refreshUnreadCount();
    }
  }

  toggleTheme(): void {
    this.themeService.toggle();
  }

  toggleMenu(): void {
    this.menuOpen = !this.menuOpen;
  }

  closeMenu(): void {
    this.menuOpen = false;
  }

  logout(): void {
    this.closeMenu();
    this.notificationService.unreadCount.set(0);
    this.authService.logout();
  }
}
