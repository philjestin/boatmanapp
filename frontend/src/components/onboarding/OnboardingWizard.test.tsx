import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { OnboardingWizard } from './OnboardingWizard';

describe('OnboardingWizard', () => {
  const mockOnComplete = vi.fn();

  beforeEach(() => {
    mockOnComplete.mockClear();
  });

  it('renders nothing when isOpen is false', () => {
    const { container } = render(
      <OnboardingWizard isOpen={false} onComplete={mockOnComplete} />
    );
    expect(container.firstChild).toBeNull();
  });

  it('renders welcome step when isOpen is true', () => {
    render(<OnboardingWizard isOpen={true} onComplete={mockOnComplete} />);

    expect(screen.getByText('Welcome to Boatman')).toBeInTheDocument();
    expect(screen.getByText('Get Started')).toBeInTheDocument();
  });

  it('navigates to approval step when clicking Get Started', () => {
    render(<OnboardingWizard isOpen={true} onComplete={mockOnComplete} />);

    fireEvent.click(screen.getByText('Get Started'));

    expect(screen.getByText('Choose Your Approval Mode')).toBeInTheDocument();
    expect(screen.getByText('Suggest Mode')).toBeInTheDocument();
    expect(screen.getByText('Auto-Edit Mode')).toBeInTheDocument();
    expect(screen.getByText('Full Auto Mode')).toBeInTheDocument();
  });

  it('can select approval mode', () => {
    render(<OnboardingWizard isOpen={true} onComplete={mockOnComplete} />);

    // Navigate to approval step
    fireEvent.click(screen.getByText('Get Started'));

    // Click on Full Auto Mode
    fireEvent.click(screen.getByText('Full Auto Mode'));

    // Continue to next step
    fireEvent.click(screen.getByText('Continue'));

    // Should be on theme step
    expect(screen.getByText('Choose Your Theme')).toBeInTheDocument();
  });

  it('can navigate through all steps and complete', () => {
    render(<OnboardingWizard isOpen={true} onComplete={mockOnComplete} />);

    // Step 1: Welcome
    fireEvent.click(screen.getByText('Get Started'));

    // Step 2: Approval Mode
    expect(screen.getByText('Choose Your Approval Mode')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Continue'));

    // Step 3: Theme
    expect(screen.getByText('Choose Your Theme')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Continue'));

    // Step 4: Notifications
    expect(screen.getByText('Enable Notifications?')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Continue'));

    // Step 5: Complete
    expect(screen.getByText("You're all set!")).toBeInTheDocument();

    // Click Start Using Boatman
    fireEvent.click(screen.getByText('Start Using Boatman'));

    // onComplete should be called with preferences
    expect(mockOnComplete).toHaveBeenCalledTimes(1);
    expect(mockOnComplete).toHaveBeenCalledWith(expect.objectContaining({
      approvalMode: 'suggest', // default
      theme: 'dark', // default
      notificationsEnabled: true, // default
      onboardingCompleted: true,
    }));
  });

  it('can navigate back through steps', () => {
    render(<OnboardingWizard isOpen={true} onComplete={mockOnComplete} />);

    // Navigate forward
    fireEvent.click(screen.getByText('Get Started'));
    fireEvent.click(screen.getByText('Continue'));

    // Should be on theme step
    expect(screen.getByText('Choose Your Theme')).toBeInTheDocument();

    // Navigate back
    fireEvent.click(screen.getByText('Back'));

    // Should be back on approval step
    expect(screen.getByText('Choose Your Approval Mode')).toBeInTheDocument();
  });

  it('displays progress bar correctly', () => {
    render(<OnboardingWizard isOpen={true} onComplete={mockOnComplete} />);

    // Step 1 of 5
    expect(screen.getByText('Step 1 of 5')).toBeInTheDocument();
    expect(screen.getByText('20% complete')).toBeInTheDocument();

    // Navigate to step 2
    fireEvent.click(screen.getByText('Get Started'));
    expect(screen.getByText('Step 2 of 5')).toBeInTheDocument();
    expect(screen.getByText('40% complete')).toBeInTheDocument();
  });
});
