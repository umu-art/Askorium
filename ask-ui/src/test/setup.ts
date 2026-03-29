import '@testing-library/jest-dom'

// jsdom doesn't implement matchMedia — provide a minimal stub
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockReturnValue({ matches: false }),
})
