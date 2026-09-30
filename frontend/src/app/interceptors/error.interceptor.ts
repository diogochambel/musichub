import { HttpContextToken, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, throwError } from 'rxjs';
import { ErrorPopupService } from '../services/error-popup.service';

export const BYPASS_GLOBAL_ERROR = new HttpContextToken<boolean>(() => false);

export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const errorPopup = inject(ErrorPopupService);
  return next(req).pipe(
    catchError((err) => {
      if (!req.context.get(BYPASS_GLOBAL_ERROR)) {
        const msg = err.error?.message || err.message || 'An unexpected error occurred';
        errorPopup.showError(msg);
      }
      return throwError(() => err);
    }),
  );
};
