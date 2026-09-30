import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { ErrorPopupComponent } from './components/error-popup/error-popup.component';
import { HeaderComponent } from './components/header/header.component';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, ErrorPopupComponent, HeaderComponent],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css'],
})
export class AppComponent {}
