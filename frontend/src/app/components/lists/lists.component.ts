import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';

import {
  CustomListService,
  MusicList,
} from '../../services/custom_list.service';

@Component({
  selector: 'app-lists',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './lists.component.html',
  styleUrls: ['./lists.component.css'],
})
export class ListsComponent implements OnInit {
  private customListService = inject(CustomListService);

  lists = signal<MusicList[]>([]);
  loading = signal(true);
  listsError = signal('');

  showForm = signal(false);
  formName = '';
  formError = '';
  formSuccess = '';
  submitting = false;

  sortField = signal<'name' | 'album_count' | 'updated_at'>('updated_at');
  sortDir = signal<'asc' | 'desc'>('desc');

  sortedLists = computed(() => {
    const data = [...this.lists()];
    const field = this.sortField();
    const dir = this.sortDir();

    return data.sort((a, b) => {
      let valA: string | number = a[field];
      let valB: string | number = b[field];

      if (field === 'name') {
        valA = String(valA).toLowerCase();
        valB = String(valB).toLowerCase();
      }

      if (field === 'updated_at') {
        valA = new Date(String(valA)).getTime();
        valB = new Date(String(valB)).getTime();
      }

      if (valA < valB) return dir === 'asc' ? -1 : 1;
      if (valA > valB) return dir === 'asc' ? 1 : -1;
      return 0;
    });
  });

  ngOnInit(): void {
    this.loadLists();
  }

  loadLists(): void {
    this.loading.set(true);
    this.listsError.set('');

    this.customListService.getLists().subscribe({
      next: (data) => {
        this.lists.set(data ?? []);
        this.loading.set(false);
      },
      error: (err) => {
        this.listsError.set(err.error?.message || 'Failed to load lists');
        this.lists.set([]);
        this.loading.set(false);
      },
    });
  }

  setSort(field: 'name' | 'album_count' | 'updated_at'): void {
    if (this.sortField() === field) {
      this.sortDir.set(this.sortDir() === 'asc' ? 'desc' : 'asc');
    } else {
      this.sortField.set(field);
      this.sortDir.set('asc');
    }
  }

  sortIcon(field: 'name' | 'album_count' | 'updated_at'): string {
    if (this.sortField() !== field) return '↕';
    return this.sortDir() === 'asc' ? '↑' : '↓';
  }

  toggleForm(): void {
    this.showForm.update((value) => !value);
    this.formError = '';
    this.formSuccess = '';
  }

  onSubmit(): void {
    this.formError = '';
    this.formSuccess = '';

    const name = this.formName.trim();

    if (!name) {
      this.formError = 'List name is required';
      return;
    }

    if (name.length > 100) {
      this.formError = 'List name must not exceed 100 characters';
      return;
    }

    this.submitting = true;

    this.customListService.createList(name).subscribe({
      next: () => {
        this.formSuccess = 'List created successfully!';
        this.submitting = false;
        this.formName = '';
        this.loadLists();

        setTimeout(() => {
          this.showForm.set(false);
          this.formSuccess = '';
        }, 1500);
      },
      error: (err) => {
        this.formError = err.error?.message || 'Failed to create list';
        this.submitting = false;
      },
    });
  }

  onDelete(list: MusicList): void {
    if (!confirm(`Are you sure you want to delete "${list.name}"?`)) {
      return;
    }

    this.customListService.deleteList(list.id).subscribe({
      next: () => {
        this.loadLists();
      },
      error: (err) => {
        this.listsError.set(err.error?.message || 'Failed to delete list');
      },
    });
  }

  formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-GB', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  }
}