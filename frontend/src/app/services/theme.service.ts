import { Injectable, signal } from '@angular/core';

type Theme = 'light' | 'dark';

const THEME_KEY = 'psi-theme';

function getInitialTheme(): Theme {
  const saved = localStorage.getItem(THEME_KEY) as Theme | null;
  if (saved === 'light' || saved === 'dark') {
    return saved;
  }
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    return 'dark';
  }
  return 'light';
}

@Injectable({ providedIn: 'root' })
export class ThemeService {
  theme = signal<Theme>(getInitialTheme());

  constructor() {
    this.applyTheme(this.theme());
  }

  toggle(): void {
    const next = this.theme() === 'light' ? 'dark' : 'light';
    this.theme.set(next);
    localStorage.setItem(THEME_KEY, next);
    this.applyTheme(next);
  }

  private applyTheme(t: Theme): void {
    document.documentElement.setAttribute('data-theme', t);
  }
}
