import { Component, inject, OnInit } from '@angular/core';
import { RouterLink } from '@angular/router';
import { CollectionService, CollectionItemResponse } from '../../services/collection.service';
import { ErrorPopupService } from '../../services/error-popup.service';

@Component({
  selector: 'app-collection',
  imports: [RouterLink],
  templateUrl: './collection.component.html',
  styleUrls: ['./collection.component.css'],
})
export class CollectionComponent implements OnInit {
  private collectionService = inject(CollectionService);
  private errorPopup = inject(ErrorPopupService);

  items: CollectionItemResponse[] = [];
  sortField = 'added_at';
  sortOrder: 'asc' | 'desc' = 'desc';
  successMessage = '';

  ngOnInit() {
    this.collectionService.getMyCollection().subscribe({
      next: (data) => (this.items = data),
      error: () => this.errorPopup.showError('Error loading collection'),
    });
  }

  get sortedItems(): CollectionItemResponse[] {
    return [...this.items].sort((a, b) => {
      const aVal = (a as any)[this.sortField] ?? '';
      const bVal = (b as any)[this.sortField] ?? '';
      const cmp = aVal < bVal ? -1 : aVal > bVal ? 1 : 0;
      return this.sortOrder === 'asc' ? cmp : -cmp;
    });
  }

  sortBy(field: string) {
    if (this.sortField === field) {
      this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
    } else {
      this.sortField = field;
      this.sortOrder = 'asc';
    }
  }

  sortIndicator(field: string): string {
    if (this.sortField !== field) return '';
    return this.sortOrder === 'asc' ? ' ▲' : ' ▼';
  }

  remove(id: string) {
    if (!confirm('Remove this item from the collection?')) return;
    this.collectionService.removeFromCollection(id).subscribe({
      next: () => {
        this.items = this.items.filter((i) => i.id !== id);
        this.successMessage = 'Item removed from collection successfully.';
        setTimeout(() => (this.successMessage = ''), 3000);
      },
      error: () => this.errorPopup.showError('Error removing item'),
    });
  }

  formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-GB');
  }

  supportLabel(support: string): string {
    if (support === 'vinyl') return 'Vinyl';
    if (support === 'cassette') return 'Cassette';
    return support;
  }
}
