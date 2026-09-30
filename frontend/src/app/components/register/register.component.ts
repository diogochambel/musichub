import { Component, inject } from '@angular/core';
import { Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [FormsModule, CommonModule],
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css'],
})
export class RegisterComponent {
  username = '';
  email = '';
  password = '';
  birthDate = '';

  errors: string[] = [];
  successMessage = '';

  private authService = inject(AuthService);
  private router = inject(Router);

  validate(): string[] {
    const errors: string[] = [];

    if (!this.username) {
      errors.push('Username is required');
    } else if (!/^[a-zA-Z0-9]+$/.test(this.username)) {
      errors.push('Username must contain only letters and numbers');
    }

    if (!this.email) {
      errors.push('Email is required');
    } else if (!/^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$/.test(this.email)) {
      errors.push('Invalid email format');
    }

    if (!this.password) {
      errors.push('Password is required');
    } else {
      if (this.password.length < 8) {
        errors.push('Password must be at least 8 characters');
      }
      if (!/[A-Z]/.test(this.password)) {
        errors.push('Password must contain at least one uppercase letter');
      }
      if (!/[a-z]/.test(this.password)) {
        errors.push('Password must contain at least one lowercase letter');
      }
      if (!/[0-9]/.test(this.password)) {
        errors.push('Password must contain at least one number');
      }
    }

    if (!this.birthDate) {
      errors.push('Birth date is required');
    } else {
      const birth = new Date(this.birthDate);
      const today = new Date();
      let age = today.getFullYear() - birth.getFullYear();
      if (today < new Date(today.getFullYear(), birth.getMonth(), birth.getDate())) {
        age--;
      }
      if (age < 13) {
        errors.push('You must be at least 13 years old');
      }
    }

    return errors;
  }

  onRegister() {
    this.errors = [];
    this.successMessage = '';

    const validationErrors = this.validate();
    if (validationErrors.length > 0) {
      this.errors = validationErrors;
      return;
    }

    this.authService
      .register(this.username, this.email, this.password, new Date(this.birthDate).toISOString())
      .subscribe({
        next: (data) => {
          if (data.token) {
            this.authService.handleLoginSuccess(data.token);
            this.successMessage = data.message || 'Account created successfully!';
            setTimeout(() => this.router.navigate(['/dashboard']), 2000);
          } else {
            this.errors = [data.message || 'Error creating account'];
          }
        },
        error: (err) => {
          const msg = err.error?.message || 'Could not connect to the server';
          this.errors = [msg];
        },
      });
  }

  goBack() {
    this.router.navigate(['/']);
  }
}
