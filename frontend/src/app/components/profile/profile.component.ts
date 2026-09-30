import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import {
  FormBuilder,
  ReactiveFormsModule,
  Validators,
  AbstractControl,
  ValidationErrors,
} from '@angular/forms';
import { RouterLink } from '@angular/router';
import { ProfileService } from '../../services/profile.service';
import { MusicService } from '../../services/music.service';
import { User, UpdateUserRequest } from '../../models/user.interface';

function usernameValidator(control: AbstractControl): ValidationErrors | null {
  const value = control.value ?? '';
  return /^[A-Za-z0-9]+$/.test(value) ? null : { invalidUsername: true };
}

function strongPasswordValidator(control: AbstractControl): ValidationErrors | null {
  const value = control.value ?? '';

  if (!value) return null;

  const valid = value.length >= 8 && /[A-Z]/.test(value) && /[a-z]/.test(value) && /\d/.test(value);

  return valid ? null : { weakPassword: true };
}

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './profile.component.html',
  styleUrl: './profile.component.css',
})
export class ProfileComponent implements OnInit {
  private fb = inject(FormBuilder);
  private profileService = inject(ProfileService);
  private musicService = inject(MusicService);

  loading = signal(true);
  saving = signal(false);
  editMode = signal(false);
  successMessage = signal('');
  profile = signal<User | null>(null);
  favoriteArtistName = signal('');

  form = this.fb.nonNullable.group({
    username: ['', [Validators.required, usernameValidator]],
    email: ['', [Validators.required, Validators.email]],
    birth_date: ['', [Validators.required]],
    current_password: ['', [Validators.required]],
    password: ['', [strongPasswordValidator]],
  });

  favoriteArtistId = computed(() => this.profile()?.favorite_artist_id ?? null);

  ngOnInit(): void {
    this.loadProfile();
  }

  loadProfile(): void {
    this.loading.set(true);
    this.successMessage.set('');
    this.favoriteArtistName.set('');

    const request = this.profileService.fetchProfile();

    if (!request) {
      this.loading.set(false);
      return;
    }

    request.subscribe({
      next: (user: User) => {
        this.profile.set(user);

        if (user.favorite_artist_id) {
          this.musicService.getArtist(user.favorite_artist_id).subscribe({
            next: (artist) => this.favoriteArtistName.set(artist.name),
            error: () => this.favoriteArtistName.set(''),
          });
        }

        this.form.patchValue({
          username: user.username,
          email: user.email,
          birth_date: this.toDateInput(user.birth_date),
          current_password: '',
          password: '',
        });

        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
      },
    });
  }

  enableEdit(): void {
    this.editMode.set(true);
    this.successMessage.set('');
  }

  cancelEdit(): void {
    this.editMode.set(false);
    this.successMessage.set('');

    const user = this.profile();
    if (!user) return;

    this.form.patchValue({
      username: user.username,
      email: user.email,
      birth_date: this.toDateInput(user.birth_date),
      current_password: '',
      password: '',
    });

    this.form.markAsPristine();
    this.form.markAsUntouched();
  }

  save(): void {
    this.successMessage.set('');

    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const { username, email, birth_date, current_password, password } = this.form.getRawValue();

    this.saving.set(true);

    const payload: UpdateUserRequest = {
      username,
      email,
      birth_date: new Date(birth_date).toISOString(),
      current_password,
      ...(password ? { password } : {}),
    };

    const request = this.profileService.updateProfile(payload);

    if (!request) {
      this.saving.set(false);
      return;
    }

    request.subscribe({
      next: (response: User) => {
        const passwordChanged = !!password;

        this.profile.set(response);
        this.editMode.set(false);
        this.saving.set(false);

        if (passwordChanged) {
          localStorage.clear();
          sessionStorage.clear();
          window.location.href = '/login';
          return;
        }

        this.successMessage.set('Profile updated successfully.');
        this.form.patchValue({
          current_password: '',
          password: '',
        });
      },
      error: (err) => {
        if (err.status === 401) {
          this.successMessage.set('Incorrect current password.');
        } else {
          this.successMessage.set('Failed to update profile.');
        }
        this.saving.set(false);
      },
    });
  }

  private toDateInput(value: string): string {
    if (!value || value.startsWith('0001-')) return '';
    return value.split('T')[0];
  }
}
