import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { environment } from '../../environments/environment';

export interface VersionRequest {
  id: string;
  album_id: string;
  album_title: string;
  ean13: string;
  support: 'CD' | 'vinyl' | 'cassette';
  version_name: string;
  status: 'em análise' | 'aceite' | 'recusado';
  requested_at: string;
}

export interface CreateRequestBody {
  album_id: string;
  ean13: string;
  support: string;
  version_name: string;
}

@Injectable({ providedIn: 'root' })
export class RequestService {
  private http = inject(HttpClient);
  private apiUrl = environment.apiUrl;

  // US12: send a new version request to the backend
  createRequest(body: CreateRequestBody) {
    return this.http.post<{ status: string; message: string; request: VersionRequest }>(
      `${this.apiUrl}/api/requests`,
      body,
    );
  }

  // US13: get requests, optionally filtered by status
  getRequests(status?: string) {
    let params = new HttpParams();
    if (status) {
      params = params.set('status', status);
    }
    return this.http.get<VersionRequest[]>(`${this.apiUrl}/api/requests`, { params });
  }

  // US14: simulate admin responding to a request
  respondToRequest(requestId: string, status: 'aceite' | 'recusado') {
    return this.http.post<{ status: string; message: string }>(
      `${this.apiUrl}/api/requests/${requestId}/respond`,
      { status },
    );
  }
}
