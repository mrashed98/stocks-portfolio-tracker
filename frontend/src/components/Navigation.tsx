import { Link, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { 
  BarChart3, 
  Target, 
  Briefcase, 
  TrendingUp,
  LineChart
} from 'lucide-react'

const navigation = [
  { name: 'Dashboard', href: '/', icon: BarChart3 },
  { name: 'Strategies', href: '/strategies', icon: Target },
  { name: 'Portfolios', href: '/portfolios', icon: Briefcase },
  { name: 'Stocks', href: '/stocks', icon: TrendingUp },
  { name: 'Charts', href: '/charts', icon: LineChart },
]

export function Navigation() {
  const location = useLocation()

  return (
    <div className="w-64 bg-card border-r border-border">
      <div className="p-6">
        <h1 className="text-2xl font-bold text-foreground">Portfolio App</h1>
      </div>
      <nav className="px-4 space-y-2">
        {navigation.map((item) => {
          const Icon = item.icon
          const isActive = location.pathname === item.href
          
          return (
            <Link
              key={item.name}
              to={item.href}
              className={cn(
                'flex items-center px-3 py-2 rounded-md text-sm font-medium transition-colors',
                isActive
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:text-foreground hover:bg-accent'
              )}
            >
              <Icon className="mr-3 h-5 w-5" />
              {item.name}
            </Link>
          )
        })}
      </nav>
    </div>
  )
}