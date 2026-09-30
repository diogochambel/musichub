import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RequestService, VersionRequest, CreateRequestBody } from '../../services/request.service';
import { MusicService } from '../../services/music.service';

@Component({
  selector: 'app-requests',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './requests.component.html',
  styleUrls: ['./requests.component.css'],
})
export class RequestsComponent implements OnInit {
  private requestService = inject(RequestService);
  private musicService = inject(MusicService);

  // --- US13: list state ---
  requests: VersionRequest[] = [];
  loadingRequests = false;
  requestsError = '';
  statusFilter = '';

  // --- US12: form state ---
  showForm = false;
  formAlbumId = '';
  formEan13 = '';
  formSupport = 'CD';
  formVersionName = '';
  formError = '';
  formSuccess = '';
  submitting = false;

  // --- Album autocomplete ---
  albumQuery = '';
  albumSuggestions: any[] = [];
  selectedAlbumLabel = '';
  searchingAlbums = false;
  private searchTimeout: any;

  ngOnInit() {
    this.loadRequests();
  }

  loadRequests() {
    this.loadingRequests = true;
    this.requestsError = '';

    this.requestService.getRequests(this.statusFilter || undefined).subscribe({
      next: (data) => {
        this.requests = data;
        this.loadingRequests = false;
      },
      error: (err) => {
        this.requestsError = err.error?.message || 'Failed to load requests';
        this.loadingRequests = false;
      },
    });
  }

  onFilterChange() {
    this.loadRequests();
  }

  toggleForm() {
    this.showForm = !this.showForm;
    this.formError = '';
    this.formSuccess = '';
  }

  onAlbumSearch() {
    clearTimeout(this.searchTimeout);
    this.albumSuggestions = [];
    const query = this.albumQuery.trim();

    if (!query || query.length < 1) {
      this.searchingAlbums = false;
      return;
    }

    this.searchingAlbums = true;
    this.searchTimeout = setTimeout(() => {
      this.musicService.searchAlbums(query).subscribe({
        next: (data) => {
          this.albumSuggestions = data.slice(0, 8);
          this.searchingAlbums = false;
        },
        error: () => {
          this.albumSuggestions = [];
          this.searchingAlbums = false;
        },
      });
    }, 300);
  }

  selectAlbum(album: any) {
    this.formAlbumId = album.id;
    this.selectedAlbumLabel = album.title + (album.artist_name ? ' — ' + album.artist_name : '');
    this.albumQuery = '';
    this.albumSuggestions = [];
  }

  clearAlbumSelection() {
    this.formAlbumId = '';
    this.selectedAlbumLabel = '';
    this.albumQuery = '';
    this.albumSuggestions = [];
  }

  onSubmit() {
    this.formError = '';
    this.formSuccess = '';

    if (!this.formAlbumId) {
      this.formError = 'Please select an album';
      return;
    }
    if (!this.formEan13 || this.formEan13.length !== 13 || !/^\d+$/.test(this.formEan13)) {
      this.formError = 'EAN-13 must be exactly 13 digits';
      return;
    }

    this.submitting = true;

    const body: CreateRequestBody = {
      album_id: this.formAlbumId,
      ean13: this.formEan13,
      support: this.formSupport,
      version_name: this.formVersionName,
    };

    this.requestService.createRequest(body).subscribe({
      next: (data) => {
        this.formSuccess = data.message || 'Request submitted successfully!';
        this.submitting = false;
      this.formAlbumId = '';
      this.selectedAlbumLabel = '';
      this.albumQuery = '';
      this.albumSuggestions = [];
      this.formEan13 = '';
      this.formSupport = 'CD';
      this.formVersionName = '';
      this.loadRequests();
        setTimeout(() => {
          this.showForm = false;
          this.formSuccess = '';
        }, 2000);
      },
      error: (err) => {
        this.formError = err.error?.message || 'Failed to submit request';
        this.submitting = false;
      },
    });
  }

  getStatusClass(status: string): string {
    switch (status) {
      case 'aceite':
        return 'status-accepted';
      case 'recusado':
        return 'status-rejected';
      default:
        return 'status-pending';
    }
  }

  statusLabel(status: string): string {
    switch (status) {
      case 'aceite':
        return 'Accepted';
      case 'recusado':
        return 'Rejected';
      default:
        return 'Pending';
    }
  }

  formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-GB', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  }
}
