import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';

import {
  CustomListService,
  ListItem,
} from '../../services/custom_list.service';

@Component({
  selector: 'app-list-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './list_detail.component.html',
  styleUrls: ['./list_detail.component.css'],
})
export class ListDetailComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private customListService = inject(CustomListService);

  listId = '';
  listName = signal('Custom List');

  items = signal<ListItem[]>([]);
  loading = signal(true);
  error = signal('');
  success = signal('');

  sortField = signal<'album_title' | 'added_at'>('added_at');
  sortDir = signal<'asc' | 'desc'>('desc');

  sortedItems = computed(() => {
    const data = [...this.items()];
    const field = this.sortField();
    const dir = this.sortDir();

    return data.sort((a, b) => {
      let valA: string | number;
      let valB: string | number;

      if (field === 'album_title') {
        valA = this.albumTitle(a).toLowerCase();
        valB = this.albumTitle(b).toLowerCase();
      } else {
        valA = new Date(a.added_at).getTime();
        valB = new Date(b.added_at).getTime();
      }

      if (valA < valB) return dir === 'asc' ? -1 : 1;
      if (valA > valB) return dir === 'asc' ? 1 : -1;
      return 0;
    });
  });

  ngOnInit(): void {
    this.listId = this.route.snapshot.paramMap.get('id') || '';

    this.loadListName();
    this.loadItems();
  }

  loadListName(): void {
    if (!this.listId) return;

    this.customListService.getLists().subscribe({
      next: (lists) => {
        const list = lists.find((item) => item.id === this.listId);

        if (list) {
          this.listName.set(list.name);
        }
      },
    });
  }

  loadItems(): void {
    if (!this.listId) {
      this.error.set('Invalid list id');
      this.loading.set(false);
      return;
    }

    this.loading.set(true);
    this.error.set('');

    this.customListService.getListItems(this.listId).subscribe({
      next: (items) => {
        this.items.set(items ?? []);
        this.loading.set(false);
      },
      error: (err) => {
        this.error.set(err.error?.message || 'Failed to load list');
        this.items.set([]);
        this.loading.set(false);
      },
    });
  }

  setSort(field: 'album_title' | 'added_at'): void {
    if (this.sortField() === field) {
      this.sortDir.set(this.sortDir() === 'asc' ? 'desc' : 'asc');
    } else {
      this.sortField.set(field);
      this.sortDir.set('asc');
    }
  }

  sortIcon(field: 'album_title' | 'added_at'): string {
    if (this.sortField() !== field) return '↕';
    return this.sortDir() === 'asc' ? '↑' : '↓';
  }

  removeAlbum(item: ListItem): void {
    this.customListService.removeAlbumFromList(this.listId, item.album_id).subscribe({
      next: () => {
        this.success.set('Album removed from list');

        this.items.update((items) =>
          items.filter((current) => current.album_id !== item.album_id),
        );

        setTimeout(() => this.success.set(''), 1500);
      },
      error: (err) => {
        this.error.set(err.error?.message || 'Failed to remove album');
      },
    });
  }

  albumTitle(item: ListItem): string {
    return item.album_title || item.title || 'Unknown album';
  }

  formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-GB', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  }
}