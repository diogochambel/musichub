import { Component, inject } from '@angular/core';
import { ErrorPopupService } from '../../services/error-popup.service';

@Component({
  selector: 'app-error-popup',
  templateUrl: './error-popup.component.html',
  styleUrls: ['./error-popup.component.css'],
})
export class ErrorPopupComponent {
  readonly errorPopup = inject(ErrorPopupService);
  readonly ANIMATION_DURATION_MS = 250;

  isClosing = false;
  private closeTimer: ReturnType<typeof setTimeout> | null = null;

  onBackdropClick() {
    this.startClose();
  }

  onPopupClick(event: Event) {
    event.stopPropagation();
  }

  onCloseClick() {
    this.startClose();
  }

  private startClose() {
    if (this.isClosing) return;
    this.isClosing = true;
    this.closeTimer = setTimeout(() => {
      this.errorPopup.dismiss();
      this.isClosing = false;
      this.closeTimer = null;
    }, this.ANIMATION_DURATION_MS);
  }
}
