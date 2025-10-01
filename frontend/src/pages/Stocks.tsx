import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import TradingViewChart from '@/components/TradingViewChart'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import { LoadingState } from '@/components/ui/loading-spinner'
import { useQuery } from '@tanstack/react-query'
import { apiClient } from '@/lib/api-client'
// import { useToast } from '@/hooks/use-toast'
import { Search, TrendingUp, TrendingDown, Activity } from 'lucide-react'

interface Stock {
  id: string
  ticker: string
  name: string
  sector?: string
  current_price?: number
  change?: number
  change_percent?: number
  signal?: 'Buy' | 'Hold'
  last_updated?: string
}

type StockView = 'list' | 'detail'

export default function Stocks() {
  const [currentView, setCurrentView] = useState<StockView>('list')
  const [selectedStock, setSelectedStock] = useState<Stock | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const navigate = useNavigate()
  // const { toast } = useToast()

  // Fetch stocks
  const { data: stocks = [], isLoading, error } = useQuery({
    queryKey: ['stocks', searchQuery],
    queryFn: async () => {
      const params = searchQuery ? `?search=${encodeURIComponent(searchQuery)}` : ''
      return apiClient.get<Stock[]>(`/stocks${params}`)
    },
  })

  const handleViewStock = (stock: Stock) => {
    setSelectedStock(stock)
    setCurrentView('detail')
  }

  const handleBackToList = () => {
    setCurrentView('list')
    setSelectedStock(null)
  }

  const handleNavigateToChart = (ticker: string) => {
    navigate(`/charts?symbol=${ticker}`)
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(amount)
  }

  const formatPercentage = (value: number) => {
    return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`
  }

  if (currentView === 'detail' && selectedStock) {
    return (
      <ErrorBoundary>
        <div className="space-y-6">
          <div className="flex items-center space-x-4">
            <button
              onClick={handleBackToList}
              className="p-2 text-muted-foreground hover:text-foreground"
            >
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
            </button>
            <div>
              <h1 className="text-3xl font-bold text-foreground">
                {selectedStock.ticker} - {selectedStock.name}
              </h1>
              <p className="text-muted-foreground">
                {selectedStock.sector && `${selectedStock.sector} • `}
                Stock details and chart analysis
              </p>
            </div>
          </div>

          {/* Stock Info Card */}
          <Card className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div>
                <h3 className="text-sm font-medium text-muted-foreground">Current Price</h3>
                <p className="text-2xl font-bold">
                  {selectedStock.current_price ? formatCurrency(selectedStock.current_price) : 'N/A'}
                </p>
              </div>
              <div>
                <h3 className="text-sm font-medium text-muted-foreground">Change</h3>
                <p className={`text-2xl font-bold ${
                  selectedStock.change && selectedStock.change >= 0 ? 'text-green-600' : 'text-red-600'
                }`}>
                  {selectedStock.change ? formatCurrency(selectedStock.change) : 'N/A'}
                </p>
              </div>
              <div>
                <h3 className="text-sm font-medium text-muted-foreground">Change %</h3>
                <p className={`text-2xl font-bold ${
                  selectedStock.change_percent && selectedStock.change_percent >= 0 ? 'text-green-600' : 'text-red-600'
                }`}>
                  {selectedStock.change_percent ? formatPercentage(selectedStock.change_percent) : 'N/A'}
                </p>
              </div>
              <div>
                <h3 className="text-sm font-medium text-muted-foreground">Signal</h3>
                <div className="flex items-center space-x-2">
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                    selectedStock.signal === 'Buy' 
                      ? 'bg-green-100 text-green-800' 
                      : 'bg-yellow-100 text-yellow-800'
                  }`}>
                    {selectedStock.signal || 'Hold'}
                  </span>
                </div>
              </div>
            </div>
          </Card>

          {/* TradingView Chart */}
          <Card className="p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">Price Chart</h2>
              <Button
                onClick={() => handleNavigateToChart(selectedStock.ticker)}
                variant="outline"
                size="sm"
              >
                Full Chart View
              </Button>
            </div>
            <div className="h-96">
              <TradingViewChart
                symbol={selectedStock.ticker}
                interval="1D"
                theme="light"
                height={384}
                datafeedUrl="/api/tradingview"
              />
            </div>
          </Card>
        </div>
      </ErrorBoundary>
    )
  }

  return (
    <ErrorBoundary>
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Stocks</h1>
          <p className="text-muted-foreground">
            Browse and analyze stocks in your investment universe
          </p>
        </div>

        {/* Search Bar */}
        <Card className="p-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
            <input
              type="text"
              placeholder="Search stocks by ticker or name..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
        </Card>

        {/* Stocks List */}
        {isLoading ? (
          <LoadingState message="Loading stocks..." />
        ) : error ? (
          <Card className="p-6">
            <div className="text-center text-red-600">
              Failed to load stocks. Please try again.
            </div>
          </Card>
        ) : stocks.length === 0 ? (
          <Card className="p-6">
            <div className="text-center text-muted-foreground">
              {searchQuery ? 'No stocks found matching your search.' : 'No stocks available.'}
            </div>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {stocks.map((stock) => (
              <Card key={stock.id} className="p-4 hover:shadow-md transition-shadow cursor-pointer">
                <div onClick={() => handleViewStock(stock)}>
                  <div className="flex items-center justify-between mb-2">
                    <div>
                      <h3 className="font-semibold text-lg">{stock.ticker}</h3>
                      <p className="text-sm text-muted-foreground truncate">{stock.name}</p>
                    </div>
                    <div className="text-right">
                      {stock.current_price && (
                        <p className="font-medium">{formatCurrency(stock.current_price)}</p>
                      )}
                      {stock.change_percent !== undefined && (
                        <div className={`flex items-center text-sm ${
                          stock.change_percent >= 0 ? 'text-green-600' : 'text-red-600'
                        }`}>
                          {stock.change_percent >= 0 ? (
                            <TrendingUp className="h-3 w-3 mr-1" />
                          ) : (
                            <TrendingDown className="h-3 w-3 mr-1" />
                          )}
                          {formatPercentage(stock.change_percent)}
                        </div>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                      {stock.sector && (
                        <span className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded">
                          {stock.sector}
                        </span>
                      )}
                      {stock.signal && (
                        <span className={`text-xs px-2 py-1 rounded ${
                          stock.signal === 'Buy' 
                            ? 'bg-green-100 text-green-800' 
                            : 'bg-yellow-100 text-yellow-800'
                        }`}>
                          {stock.signal}
                        </span>
                      )}
                    </div>
                    <Button
                      onClick={(e) => {
                        e.stopPropagation()
                        handleNavigateToChart(stock.ticker)
                      }}
                      variant="ghost"
                      size="sm"
                    >
                      <Activity className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </ErrorBoundary>
  )
}