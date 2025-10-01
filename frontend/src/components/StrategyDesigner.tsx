import React, { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Plus, Edit2, Trash2, Save, X } from 'lucide-react';
import { Card } from './ui/card';
import { useToast } from '../hooks/use-toast';
import { strategyService } from '../services/strategyService';
import type { 
  Strategy, 
  CreateStrategyFormData, 
  UpdateStrategyFormData
} from '../types/strategy';
import { createStrategySchema, updateStrategySchema, validateStrategyWeights } from '../types/strategy';

interface StrategyDesignerProps {
  onStrategySelect?: (strategy: Strategy) => void;
  selectedStrategyId?: string;
}



export const StrategyDesigner: React.FC<StrategyDesignerProps> = ({
  onStrategySelect,
  selectedStrategyId,
}) => {
  const [isCreating, setIsCreating] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [weightErrors, setWeightErrors] = useState<string | null>(null);
  
  const { toast } = useToast();
  const queryClient = useQueryClient();

  // Fetch strategies
  const { data: strategies = [], isLoading, error } = useQuery({
    queryKey: ['strategies'],
    queryFn: strategyService.getStrategies,
  });

  // Create strategy mutation
  const createMutation = useMutation({
    mutationFn: strategyService.createStrategy,
    onSuccess: (newStrategy) => {
      queryClient.invalidateQueries({ queryKey: ['strategies'] });
      setIsCreating(false);
      toast({
        title: 'Success',
        description: 'Strategy created successfully',
      });
      if (onStrategySelect) {
        onStrategySelect(newStrategy);
      }
    },
    onError: (error: any) => {
      toast({
        title: 'Error',
        description: error.response?.data?.details || 'Failed to create strategy',
        variant: 'destructive',
      });
    },
  });

  // Update strategy mutation
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateStrategyFormData }) =>
      strategyService.updateStrategy(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['strategies'] });
      setEditingId(null);
      toast({
        title: 'Success',
        description: 'Strategy updated successfully',
      });
    },
    onError: (error: any) => {
      toast({
        title: 'Error',
        description: error.response?.data?.details || 'Failed to update strategy',
        variant: 'destructive',
      });
    },
  });

  // Delete strategy mutation
  const deleteMutation = useMutation({
    mutationFn: strategyService.deleteStrategy,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['strategies'] });
      toast({
        title: 'Success',
        description: 'Strategy deleted successfully',
      });
    },
    onError: (error: any) => {
      toast({
        title: 'Error',
        description: error.response?.data?.details || 'Failed to delete strategy',
        variant: 'destructive',
      });
    },
  });

  // Update weight mutation
  const updateWeightMutation = useMutation({
    mutationFn: ({ id, weight }: { id: string; weight: number }) =>
      strategyService.updateStrategyWeight(id, weight),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['strategies'] });
      toast({
        title: 'Success',
        description: 'Strategy weight updated successfully',
      });
    },
    onError: (error: any) => {
      toast({
        title: 'Error',
        description: error.response?.data?.details || 'Failed to update weight',
        variant: 'destructive',
      });
    },
  });

  // Form for creating new strategy
  const createForm = useForm<CreateStrategyFormData>({
    resolver: zodResolver(createStrategySchema),
    defaultValues: {
      name: '',
      weight_mode: 'percent',
      weight_value: 0,
    },
  });

  // Form for editing strategy
  const editForm = useForm<UpdateStrategyFormData>({
    resolver: zodResolver(updateStrategySchema),
  });

  // Validate strategy weights whenever strategies change
  useEffect(() => {
    if (strategies.length > 0) {
      const error = validateStrategyWeights(strategies);
      setWeightErrors(error);
    }
  }, [strategies]);

  // Handle create strategy
  const handleCreate = (data: CreateStrategyFormData) => {
    // Check weight constraints for percentage mode
    if (data.weight_mode === 'percent') {
      const currentPercentageTotal = strategies
        .filter(s => s.weight_mode === 'percent')
        .reduce((sum, s) => sum + s.weight_value, 0);
      
      if (currentPercentageTotal + data.weight_value > 100) {
        toast({
          title: 'Validation Error',
          description: `Total percentage weights cannot exceed 100%. Current total: ${currentPercentageTotal}%, adding ${data.weight_value}% would exceed the limit.`,
          variant: 'destructive',
        });
        return;
      }
    }

    createMutation.mutate(data);
  };

  // Handle update strategy
  const handleUpdate = (id: string, data: UpdateStrategyFormData) => {
    // Check weight constraints for percentage mode
    if (data.weight_mode === 'percent' || data.weight_value) {
      const strategy = strategies.find(s => s.id === id);
      if (strategy) {
        const otherStrategies = strategies.filter(s => s.id !== id && s.weight_mode === 'percent');
        const otherTotal = otherStrategies.reduce((sum, s) => sum + s.weight_value, 0);
        const newWeight = data.weight_value || strategy.weight_value;
        const newMode = data.weight_mode || strategy.weight_mode;
        
        if (newMode === 'percent' && otherTotal + newWeight > 100) {
          toast({
            title: 'Validation Error',
            description: `Total percentage weights cannot exceed 100%. Other strategies total: ${otherTotal}%, new weight: ${newWeight}%`,
            variant: 'destructive',
          });
          return;
        }
      }
    }

    updateMutation.mutate({ id, data });
  };

  // Handle weight update
  const handleWeightUpdate = (id: string, weight: number) => {
    const strategy = strategies.find(s => s.id === id);
    if (strategy && strategy.weight_mode === 'percent') {
      const otherStrategies = strategies.filter(s => s.id !== id && s.weight_mode === 'percent');
      const otherTotal = otherStrategies.reduce((sum, s) => sum + s.weight_value, 0);
      
      if (otherTotal + weight > 100) {
        toast({
          title: 'Validation Error',
          description: `Total percentage weights cannot exceed 100%. Other strategies total: ${otherTotal}%, new weight: ${weight}%`,
          variant: 'destructive',
        });
        return;
      }
    }

    updateWeightMutation.mutate({ id, weight });
  };

  // Start editing a strategy
  const startEditing = (strategy: Strategy) => {
    setEditingId(strategy.id);
    editForm.reset({
      name: strategy.name,
      weight_mode: strategy.weight_mode,
      weight_value: strategy.weight_value,
    });
  };

  // Cancel editing
  const cancelEditing = () => {
    setEditingId(null);
    editForm.reset();
  };

  // Calculate total percentage weight
  const totalPercentageWeight = strategies
    .filter(s => s.weight_mode === 'percent')
    .reduce((sum, s) => sum + s.weight_value, 0);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-gray-500">Loading strategies...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="text-red-500">Error loading strategies</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Strategy Designer</h2>
          <p className="text-gray-600">Create and manage your investment strategies</p>
        </div>
        <button
          onClick={() => setIsCreating(true)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          <Plus className="w-4 h-4" />
          New Strategy
        </button>
      </div>

      {/* Weight validation warning */}
      {weightErrors && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
          <div className="text-red-800 font-medium">Weight Validation Error</div>
          <div className="text-red-600">{weightErrors}</div>
        </div>
      )}

      {/* Total percentage display */}
      <div className="p-4 bg-gray-50 rounded-lg">
        <div className="flex items-center justify-between">
          <span className="text-gray-700">Total Percentage Weight:</span>
          <span className={`font-bold ${totalPercentageWeight > 100 ? 'text-red-600' : 'text-green-600'}`}>
            {totalPercentageWeight.toFixed(2)}%
          </span>
        </div>
        <div className="mt-2 w-full bg-gray-200 rounded-full h-2">
          <div
            className={`h-2 rounded-full transition-all ${
              totalPercentageWeight > 100 ? 'bg-red-500' : 'bg-green-500'
            }`}
            style={{ width: `${Math.min(totalPercentageWeight, 100)}%` }}
          />
        </div>
      </div>

      {/* Create strategy form */}
      {isCreating && (
        <Card className="p-6">
          <form onSubmit={createForm.handleSubmit(handleCreate)} className="space-y-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold">Create New Strategy</h3>
              <button
                type="button"
                onClick={() => setIsCreating(false)}
                className="text-gray-500 hover:text-gray-700"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Strategy Name
              </label>
              <input
                {...createForm.register('name')}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                placeholder="Enter strategy name"
              />
              {createForm.formState.errors.name && (
                <p className="text-red-500 text-sm mt-1">{createForm.formState.errors.name.message}</p>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Weight Mode
              </label>
              <select
                {...createForm.register('weight_mode')}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="percent">Percentage</option>
                <option value="budget">Fixed Budget</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Weight Value
              </label>
              <input
                {...createForm.register('weight_value', { valueAsNumber: true })}
                type="number"
                step="0.01"
                min="0"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                placeholder="Enter weight value"
              />
              {createForm.formState.errors.weight_value && (
                <p className="text-red-500 text-sm mt-1">{createForm.formState.errors.weight_value.message}</p>
              )}
            </div>

            <div className="flex gap-2">
              <button
                type="submit"
                disabled={createMutation.isPending}
                className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
              >
                <Save className="w-4 h-4" />
                {createMutation.isPending ? 'Creating...' : 'Create Strategy'}
              </button>
              <button
                type="button"
                onClick={() => setIsCreating(false)}
                className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
              >
                Cancel
              </button>
            </div>
          </form>
        </Card>
      )}

      {/* Strategies list */}
      <div className="space-y-4">
        {strategies.map((strategy) => (
          <Card
            key={strategy.id}
            className={`p-6 transition-all ${
              selectedStrategyId === strategy.id ? 'ring-2 ring-blue-500 bg-blue-50' : 'hover:shadow-md'
            }`}
          >
            {editingId === strategy.id ? (
              // Edit form
              <form onSubmit={editForm.handleSubmit((data) => handleUpdate(strategy.id, data))} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Strategy Name
                  </label>
                  <input
                    {...editForm.register('name')}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                  {editForm.formState.errors.name && (
                    <p className="text-red-500 text-sm mt-1">{editForm.formState.errors.name.message}</p>
                  )}
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Weight Mode
                  </label>
                  <select
                    {...editForm.register('weight_mode')}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  >
                    <option value="percent">Percentage</option>
                    <option value="budget">Fixed Budget</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Weight Value
                  </label>
                  <input
                    {...editForm.register('weight_value', { valueAsNumber: true })}
                    type="number"
                    step="0.01"
                    min="0"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                  {editForm.formState.errors.weight_value && (
                    <p className="text-red-500 text-sm mt-1">{editForm.formState.errors.weight_value.message}</p>
                  )}
                </div>

                <div className="flex gap-2">
                  <button
                    type="submit"
                    disabled={updateMutation.isPending}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
                  >
                    <Save className="w-4 h-4" />
                    {updateMutation.isPending ? 'Saving...' : 'Save Changes'}
                  </button>
                  <button
                    type="button"
                    onClick={cancelEditing}
                    className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            ) : (
              // Display mode
              <div>
                <div className="flex items-center justify-between mb-4">
                  <div>
                    <h3 className="text-lg font-semibold">{strategy.name}</h3>
                    <div className="flex items-center gap-4 text-sm text-gray-600">
                      <span>Mode: {strategy.weight_mode}</span>
                      <span>
                        Weight: {strategy.weight_value}
                        {strategy.weight_mode === 'percent' ? '%' : ' USD'}
                      </span>
                      <span>Stocks: {strategy.stocks?.length || 0}</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => startEditing(strategy)}
                      className="p-2 text-gray-500 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                    >
                      <Edit2 className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => deleteMutation.mutate(strategy.id)}
                      disabled={deleteMutation.isPending}
                      className="p-2 text-gray-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors disabled:opacity-50"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>

                {/* Quick weight adjustment */}
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-600">Quick adjust weight:</span>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    defaultValue={strategy.weight_value}
                    onBlur={(e) => {
                      const newWeight = parseFloat(e.target.value);
                      if (newWeight !== strategy.weight_value && newWeight > 0) {
                        handleWeightUpdate(strategy.id, newWeight);
                      }
                    }}
                    className="w-24 px-2 py-1 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                  <span className="text-sm text-gray-600">
                    {strategy.weight_mode === 'percent' ? '%' : 'USD'}
                  </span>
                </div>

                {/* Strategy selection */}
                {onStrategySelect && (
                  <button
                    onClick={() => onStrategySelect(strategy)}
                    className={`mt-3 w-full py-2 px-4 rounded-lg transition-colors ${
                      selectedStrategyId === strategy.id
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    }`}
                  >
                    {selectedStrategyId === strategy.id ? 'Selected' : 'Select Strategy'}
                  </button>
                )}
              </div>
            )}
          </Card>
        ))}
      </div>

      {strategies.length === 0 && !isCreating && (
        <div className="text-center py-12">
          <div className="text-gray-500 mb-4">No strategies created yet</div>
          <button
            onClick={() => setIsCreating(true)}
            className="flex items-center gap-2 mx-auto px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            <Plus className="w-4 h-4" />
            Create Your First Strategy
          </button>
        </div>
      )}
    </div>
  );
};