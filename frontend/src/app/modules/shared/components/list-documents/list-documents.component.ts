import { Component, inject, input, OnChanges, OnInit, SimpleChanges } from '@angular/core';
import { Document, DocumentChunk } from '@models/document.model';
import { RequestStatus } from '@models/request-status.model';
import { DocumentService } from '@services/document.service';

@Component({
  selector: 'app-list-documents',
  imports: [],
  templateUrl: './list-documents.component.html',
  styleUrl: './list-documents.component.css'
})
export class ListDocumentsComponent implements OnInit, OnChanges {
  private readonly documentService = inject(DocumentService);

  // Documents list properties
  documents: Document[] = [];
  documentsStatus: RequestStatus = 'init';
  selectedDocumentId: string | null = null;
  chunks: DocumentChunk[] = [];
  
  // Chunks display properties
  chunksToShow: number = 3;
  readonly initialChunksToShow: number = 3;
  readonly chunksLoadIncrement: number = 5;

  // Input properties for the component
  // it is going to be a signal
  loadDocumentsOnUpload = input<boolean>(false);

  ngOnInit(): void {
    this.loadDocuments();
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['loadDocumentsOnUpload']?.currentValue) {
      this.loadDocuments();
    }
  }

  // Documents management methods
  // This method loads the list of documents from the server
  // It calls the document service to get all documents and updates the component state
  loadDocuments(): void {
    this.documentsStatus = 'loading';

    // Call the document service to get all documents
    // It subscribes to the observable returned by the service
    this.documentService.getAllDocuments().subscribe({
      next: (response) => {
        this.documents = response.documents;
        this.documentsStatus = 'success';
      },
      error: (error) => {
        this.documentsStatus = 'failed';
        console.error('Failed to load documents:', error);
      }
    });
  }

  // This method views chunks of a specific document
  // It checks if the document is already selected, and if so, it clears the selection
  viewChunks(document: Document): void {
    if (this.selectedDocumentId === document.id) {
      this.selectedDocumentId = null;
      this.chunks = [];
      this.chunksToShow = this.initialChunksToShow;
      return;
    }

    // If a different document is selected, load its chunks
    this.selectedDocumentId = document.id;
    this.chunksToShow = this.initialChunksToShow;

    // Call the document service to get chunks for the selected document
    // It subscribes to the observable returned by the service
    this.documentService.getDocumentChunks(document.id).subscribe({
      next: (response) => {
        // Update the chunks state with the response
        this.chunks = response.chunks;
      },
      error: (error) => {
        console.error('Failed to load chunks:', error);
        this.chunks = [];
      }
    });
  }

  // This method loads more chunks for display
  // It increases the number of chunks to show by the increment amount
  loadMoreChunks(): void {
    this.chunksToShow = Math.min(this.chunksToShow + this.chunksLoadIncrement, this.chunks.length);
  }

  // This method checks if there are more chunks to show
  // It returns true if there are more chunks available than currently displayed
  hasMoreChunks(): boolean {
    return this.chunks.length > this.chunksToShow;
  }

  // This method deletes a specific document
  // It prompts the user for confirmation before proceeding with the deletion
  deleteDocument(documentId: string): void {
    if (!confirm('Are you sure you want to delete this document?')) return;

    // Call the document service to delete the document
    // It subscribes to the observable returned by the service
    this.documentService.deleteDocument(documentId).subscribe({
      next: (response) => {
        console.log('Document deleted:', response);
        // Load the documents again to refresh the list
        this.loadDocuments();

        // Clear chunks if this document was selected
        if (this.selectedDocumentId === documentId) {
          this.selectedDocumentId = null;
          this.chunks = [];
        }
      },
      error: (error) => {
        console.error('Failed to delete document:', error);
        alert('Failed to delete document');
      }
    });
  }

  // Utility methods
  // These methods provide utility functions for the component
  
  // Expose Math object to template
  get Math() {
    return Math;
  }
  
  // this method gets the file icon based on the MIME type
  // It uses the document service to get the appropriate icon for the file type
  getFileIcon(mimeType: string): string {
    return this.documentService.getFileIcon(mimeType);
  }

  // This method formats the file size for display
  // It uses the document service to format the file size in a human-readable way
  formatFileSize(bytes: number): string {
    return this.documentService.formatFileSize(bytes);
  }

  // This method formats the date for display
  // It uses the document service to format the date in a human-readable way
  formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }
}
