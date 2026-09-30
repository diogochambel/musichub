import { Component, inject } from '@angular/core';
import { Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [FormsModule, CommonModule],
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css'],
})
export class LoginComponent {
  username = '';
  password = '';

  errorMessage = '';

  private authService = inject(AuthService);
  private router: Router;

  constructor(router: Router) {
    this.router = router;
  }

  onLogin() {
    this.errorMessage = '';

    if (!this.username || !this.password) {
      this.errorMessage = 'Please fill in all fields';
      return;
    }

    this.authService.login(this.username, this.password).subscribe({
      next: (data) => {
        if (data.token) {
          this.authService.handleLoginSuccess(data.token);
          this.router.navigate(['/dashboard']);
        } else {
          this.errorMessage = data.message || 'Invalid credentials';
        }
      },
      error: (err) => {
        this.errorMessage = err.error?.message || 'Could not connect to the server';
      },
    });
  }

  goBack() {
    this.router.navigate(['/']);
  }
}
