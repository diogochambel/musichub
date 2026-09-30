import { Component, OnInit, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { environment } from '../../../environments/environment';

interface HealthResponse {
  status: string;
  service?: string;
  message?: string;
}

interface DBHealthResponse {
  status: string;
  database?: string;
  message?: string;
}

@Component({
  selector: 'app-health-check',
  imports: [CommonModule],
  templateUrl: './health-check.component.html',
  styleUrls: ['./health-check.component.css'],
})
export class HealthCheckComponent implements OnInit {
  private http = inject(HttpClient);
  private apiUrl = environment.apiUrl || 'http://localhost:8080';

  backendStatus: 'checking' | 'connected' | 'error' = 'checking';
  backendMessage = 'Checking...';

  dbStatus: 'checking' | 'connected' | 'error' = 'checking';
  dbMessage = 'Checking...';

  ngOnInit() {
    this.checkBackend();
    this.checkDatabase();
  }

  private checkBackend() {
    this.http.get<HealthResponse>(`${this.apiUrl}/health`).subscribe({
      next: (res) => {
        if (res.status === 'ok') {
          this.backendStatus = 'connected';
          this.backendMessage = `Backend is running (${res.service})`;
        } else {
          this.backendStatus = 'error';
          this.backendMessage = res.message || 'Unknown error';
        }
      },
      error: (err) => {
        this.backendStatus = 'error';
        this.backendMessage = `Cannot connect to backend: ${err.message}`;
      },
    });
  }

  private checkDatabase() {
    this.http.get<DBHealthResponse>(`${this.apiUrl}/health/db`).subscribe({
      next: (res) => {
        if (res.status === 'ok') {
          this.dbStatus = 'connected';
          this.dbMessage = `Database is ${res.database}`;
        } else {
          this.dbStatus = 'error';
          this.dbMessage = res.message || 'Unknown error';
        }
      },
      error: (err) => {
        this.dbStatus = 'error';
        this.dbMessage = `Cannot check database: ${err.message}`;
      },
    });
  }

  getStatusClass(status: string): string {
    if (status === 'connected') return 'status-ok';
    if (status === 'error') return 'status-error';
    return 'status-checking';
  }
}
