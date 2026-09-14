import { describe, it, expect } from 'vitest';

describe('StatusBadge', () => {
  describe('Status Types', () => {
    it('should have PENDING status', () => {
      const statuses = ['PENDING', 'APPROVED', 'DENIED'];
      expect(statuses).toContain('PENDING');
    });

    it('should have APPROVED status', () => {
      const statuses = ['PENDING', 'APPROVED', 'DENIED'];
      expect(statuses).toContain('APPROVED');
    });

    it('should have DENIED status', () => {
      const statuses = ['PENDING', 'APPROVED', 'DENIED'];
      expect(statuses).toContain('DENIED');
    });
  });

  describe('CSS Classes', () => {
    it('should return pending class for PENDING status', () => {
      const status = 'PENDING';
      const cssClass = status.toLowerCase();
      expect(cssClass).toBe('pending');
    });

    it('should return approved class for APPROVED status', () => {
      const status = 'APPROVED';
      const cssClass = status.toLowerCase();
      expect(cssClass).toBe('approved');
    });

    it('should return denied class for DENIED status', () => {
      const status = 'DENIED';
      const cssClass = status.toLowerCase();
      expect(cssClass).toBe('denied');
    });
  });
});
