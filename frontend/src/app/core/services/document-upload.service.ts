import { Injectable, signal } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class DocumentUploadService {
  // Shared signal for upload status
  private readonly _uploadSignal = signal<boolean>(false);

  // Readonly getter for the signal
  get uploadSignal() {
    return this._uploadSignal.asReadonly();
  }

  // Method to manually set signal value
  setUploadSignal(value: boolean): void {
    this._uploadSignal.set(value);
  }
}
