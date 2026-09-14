import { describe, it, expect } from 'vitest';

describe('Request List Filtering', () => {
  const mockRequests = [
    { id: '1', requester: 'John Doe', status: 'PENDING', systemResource: 'System A', accessLevel: 'READ' },
    { id: '2', requester: 'Jane Smith', status: 'APPROVED', systemResource: 'System B', accessLevel: 'WRITE' },
    { id: '3', requester: 'Bob Johnson', status: 'DENIED', systemResource: 'System C', accessLevel: 'ADMIN' },
    { id: '4', requester: 'Alice Brown', status: 'PENDING', systemResource: 'System D', accessLevel: 'READ' },
  ];

  describe('Status Filter', () => {
    it('should filter by PENDING status', () => {
      const statusFilter = 'PENDING';
      const filtered = mockRequests.filter(r => r.status === statusFilter);
      expect(filtered.length).toBe(2);
      expect(filtered.every(r => r.status === 'PENDING')).toBe(true);
    });

    it('should filter by APPROVED status', () => {
      const statusFilter = 'APPROVED';
      const filtered = mockRequests.filter(r => r.status === statusFilter);
      expect(filtered.length).toBe(1);
      expect(filtered[0].status).toBe('APPROVED');
    });

    it('should filter by DENIED status', () => {
      const statusFilter = 'DENIED';
      const filtered = mockRequests.filter(r => r.status === statusFilter);
      expect(filtered.length).toBe(1);
      expect(filtered[0].status).toBe('DENIED');
    });

    it('should return all when no status filter', () => {
      const statusFilter = '';
      const filtered = statusFilter
        ? mockRequests.filter(r => r.status === statusFilter)
        : mockRequests;
      expect(filtered.length).toBe(4);
    });
  });

  describe('Requester Search', () => {
    it('should search by requester name (case insensitive)', () => {
      const requesterSearch = 'john';
      const search = requesterSearch.toLowerCase();
      const filtered = mockRequests.filter(r =>
        r.requester.toLowerCase().includes(search)
      );
      expect(filtered.length).toBe(2);
    });

    it('should return empty when no matches', () => {
      const requesterSearch = 'xyz';
      const search = requesterSearch.toLowerCase();
      const filtered = mockRequests.filter(r =>
        r.requester.toLowerCase().includes(search)
      );
      expect(filtered.length).toBe(0);
    });

    it('should return all when search is empty', () => {
      const requesterSearch = '';
      const filtered = requesterSearch
        ? mockRequests.filter(r => r.requester.toLowerCase().includes(requesterSearch.toLowerCase()))
        : mockRequests;
      expect(filtered.length).toBe(4);
    });
  });

  describe('Combined Filters', () => {
    it('should apply both status and requester filters', () => {
      const statusFilter = 'PENDING';
      const requesterSearch = 'alice';

      let filtered = mockRequests;
      if (statusFilter) {
        filtered = filtered.filter(r => r.status === statusFilter);
      }
      if (requesterSearch) {
        const search = requesterSearch.toLowerCase();
        filtered = filtered.filter(r => r.requester.toLowerCase().includes(search));
      }

      expect(filtered.length).toBe(1);
      expect(filtered[0].requester).toBe('Alice Brown');
    });
  });

  describe('URL Query Params', () => {
    it('should generate status query param', () => {
      const statusFilter = 'PENDING';
      const query: Record<string, string> = {};
      if (statusFilter) {
        query.status = statusFilter;
      }
      expect(query.status).toBe('PENDING');
    });

    it('should generate requester query param', () => {
      const requesterSearch = 'john';
      const query: Record<string, string> = {};
      if (requesterSearch) {
        query.requester = requesterSearch;
      }
      expect(query.requester).toBe('john');
    });

    it('should not include empty filters in query', () => {
      const statusFilter = '';
      const requesterSearch = '';
      const query: Record<string, string> = {};
      if (statusFilter) {
        query.status = statusFilter;
      }
      if (requesterSearch) {
        query.requester = requesterSearch;
      }
      expect(Object.keys(query).length).toBe(0);
    });
  });
});
