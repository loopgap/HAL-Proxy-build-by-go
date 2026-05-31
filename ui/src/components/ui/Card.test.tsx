import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Card, CardHeader, CardContent, CardFooter } from './Card'

describe('Card', () => {
  it('renders children', () => {
    render(<Card>Card content</Card>)
    expect(screen.getByText('Card content')).toBeInTheDocument()
  })

  it('applies default variant classes', () => {
    render(<Card data-testid="card">Default</Card>)
    const card = screen.getByTestId('card')
    expect(card.className).toContain('bg-white')
    expect(card.className).toContain('shadow-sm')
  })

  it('applies bordered variant classes', () => {
    render(
      <Card variant="bordered" data-testid="card">
        Bordered
      </Card>,
    )
    const card = screen.getByTestId('card')
    expect(card.className).toContain('border')
    expect(card.className).toContain('border-gray-200')
  })

  it('applies elevated variant classes', () => {
    render(
      <Card variant="elevated" data-testid="card">
        Elevated
      </Card>,
    )
    const card = screen.getByTestId('card')
    expect(card.className).toContain('shadow-lg')
  })

  it('applies padding classes', () => {
    render(
      <Card padding="lg" data-testid="card">
        Large Padding
      </Card>,
    )
    const card = screen.getByTestId('card')
    expect(card.className).toContain('p-6')
  })

  it('applies hoverable class when hoverable is true', () => {
    render(
      <Card hoverable data-testid="card">
        Hoverable
      </Card>,
    )
    const card = screen.getByTestId('card')
    expect(card.className).toContain('hover:shadow-md')
    expect(card.className).toContain('cursor-pointer')
  })

  it('applies custom className', () => {
    render(
      <Card className="custom-class" data-testid="card">
        Custom
      </Card>,
    )
    expect(screen.getByTestId('card').className).toContain('custom-class')
  })
})

describe('CardHeader', () => {
  it('renders title', () => {
    render(<CardHeader title="Test Title" />)
    expect(screen.getByText('Test Title')).toBeInTheDocument()
  })

  it('renders subtitle', () => {
    render(<CardHeader title="Title" subtitle="Test Subtitle" />)
    expect(screen.getByText('Test Subtitle')).toBeInTheDocument()
  })

  it('renders action', () => {
    render(<CardHeader title="Title" action={<button>Action</button>} />)
    expect(screen.getByRole('button', { name: /action/i })).toBeInTheDocument()
  })

  it('renders children', () => {
    render(<CardHeader>Header content</CardHeader>)
    expect(screen.getByText('Header content')).toBeInTheDocument()
  })
})

describe('CardContent', () => {
  it('renders children', () => {
    render(<CardContent>Content</CardContent>)
    expect(screen.getByText('Content')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    render(
      <CardContent className="custom-class" data-testid="content">
        Content
      </CardContent>,
    )
    expect(screen.getByTestId('content').className).toContain('custom-class')
  })
})

describe('CardFooter', () => {
  it('renders children', () => {
    render(<CardFooter>Footer</CardFooter>)
    expect(screen.getByText('Footer')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    render(
      <CardFooter className="custom-class" data-testid="footer">
        Footer
      </CardFooter>,
    )
    expect(screen.getByTestId('footer').className).toContain('custom-class')
  })
})
