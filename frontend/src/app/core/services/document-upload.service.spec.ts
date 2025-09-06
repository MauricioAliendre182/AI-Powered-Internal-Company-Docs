import { TestBed } from '@angular/core/testing';
import { DocumentUploadService } from './document-upload.service';

describe('DocumentUploadService', () => {
  let service: DocumentUploadService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(DocumentUploadService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should have upload signal initialized as false', () => {
    expect(service.uploadSignal()).toBeFalse();
  });

  it('should update upload signal when setUploadSignal is called', () => {
    // Initial state
    expect(service.uploadSignal()).toBeFalse();

    // Set to true
    service.setUploadSignal(true);
    expect(service.uploadSignal()).toBeTrue();

    // Set back to false
    service.setUploadSignal(false);
    expect(service.uploadSignal()).toBeFalse();
  });

  it('should keep the signal readonly', () => {
    const uploadSignal = service.uploadSignal;

    // Verify that the signal is readonly (no set method)
    expect(() => {
      // @ts-ignore - This should fail since the signal is readonly
      uploadSignal.set(true);
    }).toThrow();
  });

  it('should maintain signal state across multiple operations', () => {
    service.setUploadSignal(true);
    expect(service.uploadSignal()).toBeTrue();

    // Value should persist
    const uploadStatus = service.uploadSignal();
    expect(uploadStatus).toBeTrue();

    // Update again
    service.setUploadSignal(false);
    expect(service.uploadSignal()).toBeFalse();
  });
});
