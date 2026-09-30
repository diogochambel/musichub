import { Injectable, signal } from '@angular/core';

@Injectable({ providedIn: 'root' })
export class ErrorPopupService {
  readonly visible = signal(false);
  readonly message = signal('');

  showError(message: string) {
    this.message.set(message);
    this.visible.set(true);
  }

  dismiss() {
    this.visible.set(false);
    this.message.set('');
  }
}
