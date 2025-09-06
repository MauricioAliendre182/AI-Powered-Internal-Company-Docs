import { ComponentFixture, TestBed } from '@angular/core/testing';

import { QuerySectionComponent } from './query-section.component';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';

describe('QuerySectionComponent', () => {
  let component: QuerySectionComponent;
  let fixture: ComponentFixture<QuerySectionComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [QuerySectionComponent],
      providers: [provideHttpClient(), provideHttpClientTesting()]
    })
    .compileComponents();

    fixture = TestBed.createComponent(QuerySectionComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
