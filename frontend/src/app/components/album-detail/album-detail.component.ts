import { Component, inject, OnDestroy, OnInit, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { Subscription } from 'rxjs';

import { MusicService } from '../../services/music.service';
import { ErrorPopupService } from '../../services/error-popup.service';
import { CollectionService } from '../../services/collection.service';
import { AuthService } from '../../services/auth.service';

import {
  CustomListService,
  MusicList,
} from '../../services/custom_list.service';

@Component({
  selector: 'app-album-detail',
  standalone: true,
  imports: [RouterLink],
  templateUrl: './album-detail.component.html',
  styleUrls: ['./album-detail.component.css'],
})
export class AlbumDetailComponent implements OnInit, OnDestroy {
  private route = inject(ActivatedRoute);
  private musicService = inject(MusicService);
  private errorPopup = inject(ErrorPopupService);
  private collectionService = inject(CollectionService);
  private customListService = inject(CustomListService);

  protected auth = inject(AuthService);

  private routeSub!: Subscription;

  album: any = null;
  private albumId = '';

  protected userCollectionEans = signal<Set<string>>(new Set());
  protected addingEan = signal<string | null>(null);
  protected addedEan = signal<string | null>(null);
  protected successMessage = signal<string>('');

  protected lists = signal<MusicList[]>([]);
  protected selectedListId = signal('');
  protected addingToList = signal(false);
  protected listSuccessMessage = signal('');
  protected listErrorMessage = signal('');

  protected showListPicker = signal(false);

  ngOnInit(): void {
    this.routeSub = this.route.paramMap.subscribe((params) => {
      const id = params.get('id');

      if (id) {
        this.albumId = id;
        this.album = null;

        this.musicService.getAlbum(id).subscribe({
          next: (data) => {
            this.album = data;

            this.loadCollection();
            this.loadLists();
          },

          error: () => {
            this.errorPopup.showError('Failed to load album');
          },
        });
      }
    });
  }

  private loadCollection(): void {
    if (!this.auth.isLoggedIn()) return;

    this.collectionService.getMyCollection().subscribe({
      next: (items) => {
        const eans = new Set<string>();

        for (const item of items) {
          if (item.album_id === this.albumId) {
            eans.add(item.ean_13);
          }
        }

        this.userCollectionEans.set(eans);
      },
    });
  }

  private loadLists(): void {
    if (!this.auth.isLoggedIn()) return;

    this.customListService.getLists().subscribe({
      next: (lists) => {
        this.lists.set(lists ?? []);

        if (lists.length > 0) {
          this.selectedListId.set(lists[0].id);
        }
      },

      error: () => {
        this.lists.set([]);
      },
    });
  }

  protected openListPicker(): void {
    this.showListPicker.set(true);

    this.listSuccessMessage.set('');
    this.listErrorMessage.set('');
  }

  protected closeListPicker(): void {
    this.showListPicker.set(false);

    this.listErrorMessage.set('');
  }

  protected addToCollection(release: any): void {
    if (!this.auth.isLoggedIn()) return;

    this.addingEan.set(release.ean13);

    this.collectionService
      .addToCollection({
        album_id: this.albumId,
        ean_13: release.ean13,
        support: release.support,
        version_name: release.version_name || '',
      })
      .subscribe({
        next: () => {
          this.userCollectionEans.update(
            (s) => new Set(s).add(release.ean13),
          );

          this.addedEan.set(release.ean13);

          this.addingEan.set(null);

          this.successMessage.set(
            `"${this.album.title}" (${this.supportLabel(release.support)}) added to your collection.`,
          );

          setTimeout(() => {
            this.addedEan.set(null);
            this.successMessage.set('');
          }, 3000);
        },

        error: (err) => {
          this.addingEan.set(null);

          if (err.status === 409) {
            this.userCollectionEans.update(
              (s) => new Set(s).add(release.ean13),
            );
          } else {
            this.errorPopup.showError(
              'Failed to add to collection',
            );
          }
        },
      });
  }

  protected addAlbumToList(): void {
    if (!this.auth.isLoggedIn()) return;

    const listId = this.selectedListId();

    if (!listId) {
      this.listErrorMessage.set('Please select a list.');
      return;
    }

    this.addingToList.set(true);

    this.listSuccessMessage.set('');
    this.listErrorMessage.set('');

    this.customListService
      .addAlbumToList(listId, this.albumId)
      .subscribe({
        next: () => {
          this.addingToList.set(false);

          this.showListPicker.set(false);

          this.listSuccessMessage.set(
            'Album added to list successfully.',
          );

          setTimeout(() => {
            this.listSuccessMessage.set('');
          }, 3000);
        },

        error: (err) => {
          this.addingToList.set(false);
          this.showListPicker.set(false);

          if (err.status === 409) {
            this.listErrorMessage.set(
              'This album is already in that list.',
            );
          } else {
            this.listErrorMessage.set(
              err.error?.message ||
                'Failed to add album to list.',
            );
          }
        },
      });
  }

  protected onSelectedListChange(event: Event): void {
    const value = (event.target as HTMLSelectElement).value;

    this.selectedListId.set(value);
  }

  ngOnDestroy(): void {
    this.routeSub?.unsubscribe();
  }

  typeLabel(type: string): string {
    if (type === 'single') return 'Single';

    if (type === 'EP') return 'EP';

    if (type === 'LP') return 'Album';

    return type;
  }

  formatDuration(seconds: number): string {
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;

    return `${m}:${s.toString().padStart(2, '0')}`;
  }

  supportLabel(support: string): string {
    if (support === 'CD') return 'CD';

    if (support === 'vinyl') return 'Vinyl';

    if (support === 'cassette') return 'Cassette';

    return support;
  }

  hasMultipleVersionsForSupport(support: string): boolean {
    if (!this.album?.releases) return false;

    return this.album.releases.filter(
      (r: any) => r.support === support,
    ).length > 1;
  }
}