import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { UploadDocumentComponent } from '@shared/components/upload-document/upload-document.component';
import { DashboardLayoutComponent } from '@shared/components/dashboard-layout/dashboard-layout.component';
import { DocumentUploadService } from '@services/document-upload.service';

@Component({
  selector: 'app-upload',
  standalone: true,
  imports: [CommonModule, UploadDocumentComponent, DashboardLayoutComponent],
  templateUrl: './upload.component.html',
  styleUrl: './upload.component.css'
})
export class UploadComponent {
  private readonly documentUploadService = inject(DocumentUploadService);

  // Method to trigger upload completion (call this when upload is successful)
  onUploadComplete(event: boolean): void {
    // Update shared service signal
    this.documentUploadService.setUploadSignal(event);

    console.log('Upload completed, signals updated');
  }

  // Method to get the shared signal value
  getSharedUploadSignal() {
    return this.documentUploadService.uploadSignal();
  }
}
