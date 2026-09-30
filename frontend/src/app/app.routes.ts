import { Routes } from '@angular/router';
import { authGuard } from './guards/auth.guard';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () => import('./components/home/home.component').then((m) => m.HomeComponent),
  },
  {
    path: 'health',
    loadComponent: () =>
      import('./components/health-check/health-check.component').then(
        (m) => m.HealthCheckComponent,
      ),
  },
  {
    path: 'dashboard',
    canActivate: [authGuard],
    loadComponent: () =>
    import('./components/dashboard/dashboard.component').then(
      (m) => m.DashboardComponent,
    ),
  },
  {
    path: 'register',
    loadComponent: () =>
      import('./components/register/register.component').then((m) => m.RegisterComponent),
  },
  {
    path: 'login',
    loadComponent: () => import('./components/login/login.component').then((m) => m.LoginComponent),
  },
  {
    path: 'artists/:id',
    loadComponent: () =>
      import('./components/artist-detail/artist-detail.component').then(
        (m) => m.ArtistDetailComponent,
      ),
  },
  {
    path: 'artists/:id/albums',
    loadComponent: () =>
      import('./components/artist-albums/artist-albums.component').then(
        (m) => m.ArtistAlbumsComponent,
      ),
  },
  {
    path: 'albums/:id',
    loadComponent: () =>
      import('./components/album-detail/album-detail.component').then(
        (m) => m.AlbumDetailComponent,
      ),
  },
  {
    path: 'songs/:id',
    loadComponent: () =>
      import('./components/song-detail/song-detail.component').then((m) => m.SongDetailComponent),
  },
  {
    path: 'profile',
    loadComponent: () =>
      import('./components/profile/profile.component').then((m) => m.ProfileComponent),
  },
  {
    path: 'requests',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./components/requests/requests.component').then((m) => m.RequestsComponent),
  },
  {
    path: 'collection',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./components/collection/collection.component').then((m) => m.CollectionComponent),
  },
  {
    path: 'notifications',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./components/notifications/notifications.component').then(
        (m) => m.NotificationsComponent,
      ),
  },
  {
    path: 'lists',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./components/lists/lists.component').then((m) => m.ListsComponent),
  },
  {
    path: 'lists/:id',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./components/list_detail/list_detail.component').then((m) => m.ListDetailComponent,),
  },
  {
    path: '**',
    redirectTo: '',
  },
];
