import axios from 'axios';

// TradingView DataFeed API implementation
export class TradingViewDataFeed {
  private baseUrl: string;

  constructor(baseUrl: string = 'http://localhost:8080/api/v1/tradingview') {
    this.baseUrl = baseUrl;
  }

  // DataFeed interface implementation
  onReady(callback: (configuration: any) => void) {
    axios.get(`${this.baseUrl}/config`)
      .then(response => {
        callback(response.data);
      })
      .catch(error => {
        console.error('Failed to get DataFeed configuration:', error);
        // Provide fallback configuration
        callback({
          supports_search: true,
          supports_group_request: true,
          supports_marks: true,
          supports_timescale_marks: false,
          supports_time: true,
          supports_quotes: true,
          supports_symbol_info: true,
          currency_codes: ['USD'],
          exchanges: [
            { value: 'NASDAQ', name: 'NASDAQ', desc: 'NASDAQ Stock Exchange' },
            { value: 'NYSE', name: 'NYSE', desc: 'New York Stock Exchange' }
          ],
          symbols_types: [
            { name: 'All types', value: '' },
            { name: 'Stock', value: 'stock' }
          ],
          supported_resolutions: ['1', '5', '15', '30', '60', '240', 'D', 'W', 'M']
        });
      });
  }

  searchSymbols(
    userInput: string,
    exchange: string,
    symbolType: string,
    onResultReadyCallback: (symbols: any[]) => void
  ) {
    const params = new URLSearchParams({
      query: userInput,
      ...(exchange && { exchange }),
      ...(symbolType && { type: symbolType }),
      limit: '30'
    });

    axios.get(`${this.baseUrl}/symbols?${params}`)
      .then(response => {
        onResultReadyCallback(response.data || []);
      })
      .catch(error => {
        console.error('Symbol search failed:', error);
        onResultReadyCallback([]);
      });
  }

  resolveSymbol(
    symbolName: string,
    onSymbolResolvedCallback: (symbolInfo: any) => void,
    _onResolveErrorCallback: (reason: string) => void
  ) {
    // For basic implementation, create symbol info from symbol name
    const symbolInfo = {
      ticker: symbolName,
      name: symbolName,
      description: symbolName,
      type: 'stock',
      session: '0930-1600',
      timezone: 'America/New_York',
      exchange: symbolName.includes(':') ? symbolName.split(':')[0] : 'NASDAQ',
      minmov: 1,
      pricescale: 100,
      has_intraday: true,
      has_no_volume: false,
      has_weekly_and_monthly: true,
      supported_resolutions: ['1', '5', '15', '30', '60', '240', 'D', 'W', 'M'],
      volume_precision: 0,
      data_status: 'streaming',
      full_name: symbolName.includes(':') ? symbolName : `NASDAQ:${symbolName}`
    };

    setTimeout(() => {
      onSymbolResolvedCallback(symbolInfo);
    }, 0);
  }

  getBars(
    symbolInfo: any,
    resolution: string,
    periodParams: any,
    onHistoryCallback: (bars: any[], meta: any) => void,
    onErrorCallback: (error: string) => void
  ) {
    const { from, to } = periodParams;
    
    const params = new URLSearchParams({
      symbol: symbolInfo.ticker || symbolInfo.name,
      resolution: resolution,
      from: from.toString(),
      to: to.toString()
    });

    axios.get(`${this.baseUrl}/history?${params}`)
      .then(response => {
        const data = response.data;
        
        if (data.s === 'no_data') {
          onHistoryCallback([], { noData: true });
          return;
        }

        if (data.s === 'error') {
          onErrorCallback(data.errmsg || 'Failed to get historical data');
          return;
        }

        if (data.s === 'ok' && data.t && data.o && data.h && data.l && data.c) {
          const bars = [];
          for (let i = 0; i < data.t.length; i++) {
            bars.push({
              time: data.t[i] * 1000, // Convert to milliseconds
              open: data.o[i],
              high: data.h[i],
              low: data.l[i],
              close: data.c[i],
              volume: data.v ? data.v[i] : 0
            });
          }
          
          onHistoryCallback(bars, { noData: false });
        } else {
          onHistoryCallback([], { noData: true });
        }
      })
      .catch(error => {
        console.error('Failed to get historical data:', error);
        onErrorCallback('Network error while fetching historical data');
      });
  }

  subscribeBars(
    symbolInfo: any,
    resolution: string,
    onRealtimeCallback: (bar: any) => void,
    subscriberUID: string,
    onResetCacheNeededCallback: () => void
  ) {
    // For real-time data, we could implement WebSocket connection
    // For now, we'll use polling as a fallback
    console.log('Subscribe to real-time data for:', symbolInfo.ticker);
    
    // Store subscription info for potential cleanup
    this.subscriptions = this.subscriptions || new Map();
    this.subscriptions.set(subscriberUID, {
      symbolInfo,
      resolution,
      onRealtimeCallback,
      onResetCacheNeededCallback
    });
  }

  unsubscribeBars(subscriberUID: string) {
    console.log('Unsubscribe from real-time data:', subscriberUID);
    
    if (this.subscriptions) {
      this.subscriptions.delete(subscriberUID);
    }
  }

  getQuotes(
    symbols: string[],
    onDataCallback: (data: any[]) => void,
    onErrorCallback: (error: string) => void
  ) {
    const params = new URLSearchParams({
      symbols: symbols.join(',')
    });

    axios.get(`${this.baseUrl}/quotes?${params}`)
      .then(response => {
        const quotes = [];
        for (const [symbol, data] of Object.entries(response.data)) {
          if (data && typeof data === 'object' && 's' in data && data.s === 'ok') {
            const quoteData = data as any;
            quotes.push({
              n: symbol,
              s: 'ok',
              v: quoteData.v || {}
            });
          }
        }
        onDataCallback(quotes);
      })
      .catch(error => {
        console.error('Failed to get quotes:', error);
        onErrorCallback('Failed to get real-time quotes');
      });
  }

  getServerTime(callback: (time: number) => void) {
    axios.get(`${this.baseUrl}/time`)
      .then(response => {
        if (response.data.s === 'ok') {
          callback(response.data.t);
        } else {
          callback(Math.floor(Date.now() / 1000));
        }
      })
      .catch(error => {
        console.error('Failed to get server time:', error);
        callback(Math.floor(Date.now() / 1000));
      });
  }

  private subscriptions?: Map<string, any>;
}

export default TradingViewDataFeed;