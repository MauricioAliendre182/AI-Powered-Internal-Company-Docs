import { TestBed } from '@angular/core/testing';
import {
  HttpTestingController,
  provideHttpClientTesting,
} from '@angular/common/http/testing';
import { DocumentService } from './document.service';
import { environment } from '@environments/environment';
import {
  DocumentsListResponse,
  DocumentUploadResponse,
  DocumentChunksResponse,
  QueryResponse,
  QueryRequest,
} from '@models/document.model';
import { provideHttpClient } from '@angular/common/http';

describe('DocumentService', () => {
  let service: DocumentService;
  let httpMock: HttpTestingController;
  const apiUrl = environment.API_URL;

  beforeEach(() => {
    TestBed.configureTestingModule({
      // providers is to make sure the service and HttpClientTestingModule are available
      // provideHttpClient is needed for Angular 15+
      // provideHttpClientTesting is needed for HttpTestingController
      providers: [
        DocumentService,
        provideHttpClient(),
        provideHttpClientTesting(),
      ],
    });
    service = TestBed.inject(DocumentService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify(); // Verify that no unmatched requests are outstanding
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('uploadDocument', () => {
    it('should upload document and return success response', () => {
      // Arrange
      const mockFormData = new FormData();
      const file = new File(['test content'], 'test.pdf', {
        type: 'application/pdf',
      });
      mockFormData.append('file', file);

      const mockResponse: DocumentUploadResponse = {
        message: 'Document uploaded successfully',
        document: {
          id: '123-456-789',
          name: 'test.pdf',
          originalFilename: 'test.pdf',
          uploadedAt: new Date().toISOString(),
        },
        chunks_created: 5,
      };

      // Act
      service.uploadDocument(mockFormData).subscribe((response) => {
        expect(response).toEqual(mockResponse);
      });

      // Assert
      const req = httpMock.expectOne(`${apiUrl}/documents`);
      expect(req.request.method).toBe('POST');
      req.flush(mockResponse);
    });

    it('should handle upload errors correctly', () => {
      // Arrange
      const mockFormData = new FormData();
      const mockErrorResponse = {
        success: false,
        message: 'Invalid file format',
      };

      // Act
      service.uploadDocument(mockFormData).subscribe({
        next: () => fail('Should have failed with error'),
        error: (error) => {
          expect(error.status).toBe(400);
          expect(error.error).toEqual(mockErrorResponse);
        },
      });

      // Assert
      const req = httpMock.expectOne(`${apiUrl}/documents`);
      req.flush(mockErrorResponse, { status: 400, statusText: 'Bad Request' });
    });
  });

  describe('getAllDocuments', () => {
    it('should return list of documents', () => {
      // Arrange
      const mockDocuments: DocumentsListResponse = {
        documents: [
          {
            id: '1',
            name: 'Document 1.pdf',
            originalFilename: 'Document 1.pdf',
            uploadedAt: new Date().toISOString(),
          },
          {
            id: '2',
            name: 'Document 2.docx',
            originalFilename: 'Document 2.docx',
            uploadedAt: new Date().toISOString(),
          },
        ],
      };

      // Act
      service.getAllDocuments().subscribe((documents) => {
        expect(documents).toEqual(mockDocuments);
        expect(documents.documents.length).toBe(2);
      });

      // Assert
      const req = httpMock.expectOne(`${apiUrl}/documents`);
      expect(req.request.method).toBe('GET');
      req.flush(mockDocuments);
    });
  });

  describe('getDocumentChunks', () => {
    it('should return chunks for a specific document', () => {
      // Arrange
      const documentId = '123-456-789';
      const mockResponse: DocumentChunksResponse = {
        documentId: 'document1',
        chunks: [
          {
            id: 'chunk1',
            documentId: 'document1',
            content: 'This is the first chunk',
            chunk_index: 0,
            size: 1024,
            contentType: 'text/plain',
            createdAt: new Date().toISOString(),
          },
          {
            id: 'chunk2',
            documentId: 'document1',
            content: 'This is the second chunk',
            chunk_index: 1,
            size: 2048,
            contentType: 'text/plain',
            createdAt: new Date().toISOString(),
          },
        ],
      };

      // Act
      service.getDocumentChunks(documentId).subscribe((response) => {
        expect(response).toEqual(mockResponse);
        expect(response.chunks.length).toBe(2);
      });

      // Assert
      const req = httpMock.expectOne(
        `${apiUrl}/documents/${documentId}/chunks`
      );
      expect(req.request.method).toBe('GET');
      req.flush(mockResponse);
    });
  });

  describe('queryDocuments', () => {
    it('should send query and return response with AI-generated answer', () => {
      // Arrange
      const mockQuery: QueryRequest = {
        question: 'What is the vacation policy?'
      };
      const mockResponse: QueryResponse = {
        question: 'What is the vacation policy?',
        answer: 'Employees get 15 vacation days per year.',
      };

      // Act
      service.queryDocuments(mockQuery).subscribe((response) => {
        expect(response).toEqual(mockResponse);
      });

      // Assert
      const req = httpMock.expectOne(`${apiUrl}/query`);
      expect(req.request.method).toBe('POST');
      expect(req.request.body).toEqual(mockQuery);
      req.flush(mockResponse);
    });
  });

  describe('deleteDocument', () => {
    it('should delete a document and return success response', () => {
      // Arrange
      const documentId = '123-456-789';
      const mockResponse = {
        success: true,
        message: 'Document deleted successfully',
      };

      // Act
      service.deleteDocument(documentId).subscribe((response) => {
        expect(response).toEqual(mockResponse);
      });

      // Assert
      const req = httpMock.expectOne(`${apiUrl}/documents/${documentId}`);
      expect(req.request.method).toBe('DELETE');
      req.flush(mockResponse);
    });
  });
});
