import { useState, FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore, type RegisterData } from '../stores/authStore';

type AccountType = 'demo' | 'live';

export function Register() {
  const navigate = useNavigate();
  const { register, isLoading, error: authError } = useAuthStore();

  // User fields
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [fullName, setFullName] = useState('');

  // Account fields
  const [accountType, setAccountType] = useState<AccountType>('demo');
  const [accountName, setAccountName] = useState('');

  // Demo account fields
  const [demoCapital, setDemoCapital] = useState(10000);

  // Live account fields
  const [binanceApiKey, setBinanceApiKey] = useState('');
  const [binanceSecretKey, setBinanceSecretKey] = useState('');
  const [binanceTestnet, setBinanceTestnet] = useState(true);

  const [errors, setErrors] = useState<Record<string, string>>({});

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    // User validation
    if (!fullName.trim()) {
      newErrors.fullName = 'Full name is required';
    }

    if (!email) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      newErrors.email = 'Please enter a valid email';
    }

    if (!password) {
      newErrors.password = 'Password is required';
    } else if (password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters';
    } else if (!/(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/.test(password)) {
      newErrors.password = 'Password must contain uppercase, lowercase, and number';
    }

    if (password !== confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    // Account validation
    if (!accountName.trim()) {
      newErrors.accountName = 'Account name is required';
    }

    if (accountType === 'demo') {
      if (demoCapital < 1000 || demoCapital > 100000) {
        newErrors.demoCapital = 'Capital must be between $1,000 and $100,000';
      }
    } else if (accountType === 'live') {
      if (!binanceApiKey.trim()) {
        newErrors.binanceApiKey = 'Binance API key is required';
      }
      if (!binanceSecretKey.trim()) {
        newErrors.binanceSecretKey = 'Binance secret key is required';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    try {
      const registerData: RegisterData = {
        email,
        password,
        full_name: fullName,
        account_type: accountType,
        account_name: accountName,
      };

      if (accountType === 'demo') {
        registerData.demo_initial_capital = demoCapital;
      } else {
        registerData.binance_api_key = binanceApiKey;
        registerData.binance_secret_key = binanceSecretKey;
        registerData.binance_testnet = binanceTestnet;
      }

      await register(registerData);
      navigate('/');
    } catch (err) {
      console.error('Registration failed:', err);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 px-4 sm:px-6 lg:px-8 py-12">
      <div className="max-w-2xl w-full space-y-8">
        {/* Header */}
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900 dark:text-white">
            Create Your Trading Account
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
            Get started with demo or live trading
          </p>
        </div>

        {/* Registration Form */}
        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          {/* Global Error Message */}
          {authError && (
            <div className="rounded-md bg-red-50 dark:bg-red-900/20 p-4">
              <div className="flex">
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-red-800 dark:text-red-200">
                    {authError}
                  </h3>
                </div>
              </div>
            </div>
          )}

          {/* User Information Section */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6 space-y-4">
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
              User Information
            </h3>

            {/* Full Name */}
            <div>
              <label htmlFor="fullName" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Full Name
              </label>
              <input
                id="fullName"
                name="fullName"
                type="text"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                className={`w-full px-3 py-2 border ${
                  errors.fullName ? 'border-red-300 dark:border-red-700' : 'border-gray-300 dark:border-gray-600'
                } rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 dark:bg-gray-700 dark:text-white sm:text-sm`}
                placeholder="John Doe"
              />
              {errors.fullName && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.fullName}</p>}
            </div>

            {/* Email */}
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Email Address
              </label>
              <input
                id="email"
                name="email"
                type="email"
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className={`w-full px-3 py-2 border ${
                  errors.email ? 'border-red-300 dark:border-red-700' : 'border-gray-300 dark:border-gray-600'
                } rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 dark:bg-gray-700 dark:text-white sm:text-sm`}
                placeholder="you@example.com"
              />
              {errors.email && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.email}</p>}
            </div>

            {/* Password */}
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Password
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="new-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className={`w-full px-3 py-2 border ${
                  errors.password ? 'border-red-300 dark:border-red-700' : 'border-gray-300 dark:border-gray-600'
                } rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 dark:bg-gray-700 dark:text-white sm:text-sm`}
                placeholder="••••••••"
              />
              {errors.password && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.password}</p>}
              <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                Must be 8+ characters with uppercase, lowercase, and number
              </p>
            </div>

            {/* Confirm Password */}
            <div>
              <label htmlFor="confirmPassword" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Confirm Password
              </label>
              <input
                id="confirmPassword"
                name="confirmPassword"
                type="password"
                autoComplete="new-password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className={`w-full px-3 py-2 border ${
                  errors.confirmPassword ? 'border-red-300 dark:border-red-700' : 'border-gray-300 dark:border-gray-600'
                } rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 dark:bg-gray-700 dark:text-white sm:text-sm`}
                placeholder="••••••••"
              />
              {errors.confirmPassword && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.confirmPassword}</p>}
            </div>
          </div>

          {/* Account Setup Section */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6 space-y-4">
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
              Trading Account Setup
            </h3>

            {/* Account Type Selection */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                Account Type
              </label>
              <div className="grid grid-cols-2 gap-4">
                {/* Demo Account */}
                <button
                  type="button"
                  onClick={() => setAccountType('demo')}
                  className={`p-4 border-2 rounded-lg text-left transition-all ${
                    accountType === 'demo'
                      ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20'
                      : 'border-gray-300 dark:border-gray-600 hover:border-gray-400'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium text-gray-900 dark:text-white">Demo Account</h4>
                      <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        Practice with virtual capital
                      </p>
                    </div>
                    <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center ${
                      accountType === 'demo' ? 'border-indigo-500' : 'border-gray-300'
                    }`}>
                      {accountType === 'demo' && <div className="w-3 h-3 rounded-full bg-indigo-500" />}
                    </div>
                  </div>
                </button>

                {/* Live Account */}
                <button
                  type="button"
                  onClick={() => setAccountType('live')}
                  className={`p-4 border-2 rounded-lg text-left transition-all ${
                    accountType === 'live'
                      ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20'
                      : 'border-gray-300 dark:border-gray-600 hover:border-gray-400'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium text-gray-900 dark:text-white">Live Account</h4>
                      <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        Trade with real funds
                      </p>
                    </div>
                    <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center ${
                      accountType === 'live' ? 'border-indigo-500' : 'border-gray-300'
                    }`}>
                      {accountType === 'live' && <div className="w-3 h-3 rounded-full bg-indigo-500" />}
                    </div>
                  </div>
                </button>
              </div>
            </div>

            {/* Account Name */}
            <div>
              <label htmlFor="accountName" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Account Name
              </label>
              <input
                id="accountName"
                name="accountName"
                type="text"
                value={accountName}
                onChange={(e) => setAccountName(e.target.value)}
                className={`w-full px-3 py-2 border ${
                  errors.accountName ? 'border-red-300 dark:border-red-700' : 'border-gray-300 dark:border-gray-600'
                } rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 dark:bg-gray-700 dark:text-white sm:text-sm`}
                placeholder="My Trading Account"
              />
              {errors.accountName && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.accountName}</p>}
            </div>

            {/* Demo Account Specific Fields */}
            {accountType === 'demo' && (
              <div>
                <label htmlFor="demoCapital" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Initial Capital: ${demoCapital.toLocaleString()}
                </label>
                <input
                  id="demoCapital"
                  name="demoCapital"
                  type="range"
                  min="1000"
                  max="100000"
                  step="1000"
                  value={demoCapital}
                  onChange={(e) => setDemoCapital(Number(e.target.value))}
                  className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer dark:bg-gray-700"
                />
                <div className="flex justify-between text-xs text-gray-500 dark:text-gray-400 mt-1">
                  <span>$1,000</span>
                  <span>$100,000</span>
                </div>
                {errors.demoCapital && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.demoCapital}</p>}
              </div>
            )}

            {/* Live Account Specific Fields */}
            {accountType === 'live' && (
              <div className="space-y-4">
                {/* Warning */}
                <div className="rounded-md bg-yellow-50 dark:bg-yellow-900/20 p-4">
                  <div className="flex">
                    <div className="ml-3">
                      <h3 className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                        Security Notice
                      </h3>
                      <p className="mt-1 text-sm text-yellow-700 dark:text-yellow-300">
                        Your API keys will be encrypted and stored securely. We recommend using Binance Testnet for initial testing.
                      </p>
                    </div>
                  </div>
                </div>

                {/* Binance API Key */}
                <div>
                  <label htmlFor="binanceApiKey" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Binance API Key
                  </label>
                  <input
                    id="binanceApiKey"
                    name="binanceApiKey"
                    type="text"
                    value={binanceApiKey}
                    onChange={(e) => setBinanceApiKey(e.target.value)}
                    className={`w-full px-3 py-2 border ${
                      errors.binanceApiKey ? 'border-red-300 dark:border-red-700' : 'border-gray-300 dark:border-gray-600'
                    } rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 dark:bg-gray-700 dark:text-white sm:text-sm font-mono`}
                    placeholder="Enter your Binance API key"
                  />
                  {errors.binanceApiKey && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.binanceApiKey}</p>}
                </div>

                {/* Binance Secret Key */}
                <div>
                  <label htmlFor="binanceSecretKey" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Binance Secret Key
                  </label>
                  <input
                    id="binanceSecretKey"
                    name="binanceSecretKey"
                    type="password"
                    value={binanceSecretKey}
                    onChange={(e) => setBinanceSecretKey(e.target.value)}
                    className={`w-full px-3 py-2 border ${
                      errors.binanceSecretKey ? 'border-red-300 dark:border-red-700' : 'border-gray-300 dark:border-gray-600'
                    } rounded-md focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 dark:bg-gray-700 dark:text-white sm:text-sm font-mono`}
                    placeholder="Enter your Binance secret key"
                  />
                  {errors.binanceSecretKey && <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.binanceSecretKey}</p>}
                </div>

                {/* Testnet Toggle */}
                <div className="flex items-center">
                  <input
                    id="binanceTestnet"
                    name="binanceTestnet"
                    type="checkbox"
                    checked={binanceTestnet}
                    onChange={(e) => setBinanceTestnet(e.target.checked)}
                    className="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded"
                  />
                  <label htmlFor="binanceTestnet" className="ml-2 block text-sm text-gray-900 dark:text-gray-300">
                    Use Binance Testnet (recommended for testing)
                  </label>
                </div>
              </div>
            )}
          </div>

          {/* Submit Button */}
          <div>
            <button
              type="submit"
              disabled={isLoading}
              className="group relative w-full flex justify-center py-3 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? (
                <span className="flex items-center">
                  <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Creating account...
                </span>
              ) : (
                'Create Account'
              )}
            </button>
          </div>

          {/* Sign in link */}
          <div className="text-center">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Already have an account?{' '}
              <Link to="/login" className="font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300">
                Sign in
              </Link>
            </p>
          </div>
        </form>
      </div>
    </div>
  );
}
