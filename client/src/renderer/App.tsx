import React, { useState, useEffect } from 'react';
import {
  ChakraProvider,
  Box,
  VStack,
  HStack,
  Heading,
  Button,
  Card,
  CardBody,
  Text,
  Badge,
  useToast,
  Spinner,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  Divider,
  Input,
  FormControl,
  FormLabel,
} from '@chakra-ui/react';

import { AuthService } from '../services/auth';
import { storageService } from '../services/storage';
import type { OAuthCallbackData, AuthState } from '../types/global';

const App: React.FC = () => {
  const [authState, setAuthState] = useState<AuthState>(storageService.getAuthState());
  const [isLoading, setIsLoading] = useState(false);
  const [serverOnline, setServerOnline] = useState<boolean | null>(null);
  const [backendURL, setBackendURL] = useState(storageService.getBackendURL());
  const toast = useToast();

  // Check server connection on mount
  useEffect(() => {
    checkServerConnection();
  }, []);

  // Set up OAuth callback listener
  useEffect(() => {
    const cleanup = window.electronAPI.onOAuthCallback(handleOAuthCallback);
    return cleanup;
  }, []);

  const checkServerConnection = async () => {
    try {
      AuthService.setBackendURL(backendURL);
      const isOnline = await AuthService.testConnection();
      setServerOnline(isOnline);
      
      if (!isOnline) {
        toast({
          title: 'Backend Offline',
          description: 'Cannot connect to the backend server. Please ensure it\'s running.',
          status: 'warning',
          duration: 5000,
          isClosable: true,
        });
      }
    } catch (error) {
      setServerOnline(false);
    }
  };

  const handleOAuthCallback = async (data: OAuthCallbackData) => {
    setIsLoading(false);
    
    if (!data.success) {
      toast({
        title: 'Authentication Failed',
        description: data.error || 'Unknown error occurred',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      return;
    }

    if (!data.code || !data.state) {
      toast({
        title: 'Authentication Failed',
        description: 'Missing authorization code or state',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      return;
    }

    try {
      let result;
      if (data.provider === 'google') {
        result = await AuthService.exchangeGoogleCode(data.code, data.state);
      } else if (data.provider === 'github') {
        // For GitHub, we need to use the current JWT token
        const currentToken = storageService.getToken();
        if (!currentToken) {
          throw new Error('Must be authenticated to link GitHub accounts');
        }
        result = await AuthService.linkGitHubAccount(data.code, data.state, currentToken);
      } else {
        throw new Error(`Unknown provider: ${data.provider}`);
      }

      if (result.success) {
        if (data.provider === 'google' && result.token) {
          // Google OAuth returns a JWT token for authentication
          storageService.storeToken(result.token);
        } else if (data.provider === 'github' && result.githubUsername) {
          // GitHub OAuth returns the linked username
          storageService.addGitHubAccount(result.githubUsername);
        }

        setAuthState(storageService.getAuthState());
        
        toast({
          title: data.provider === 'google' ? 'Authentication Successful' : 'GitHub Account Linked',
          description: data.provider === 'google' 
            ? `Successfully signed in with Google`
            : `Successfully linked GitHub account: ${result.githubUsername}`,
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      } else {
        throw new Error(result.error || 'Authentication failed');
      }
    } catch (error) {
      toast({
        title: 'Authentication Failed',
        description: (error as Error).message,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  const startOAuth = async (provider: 'google' | 'github') => {
    if (!serverOnline) {
      toast({
        title: 'Server Offline',
        description: 'Backend server is not responding. Please start the server first.',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      return;
    }

    setIsLoading(true);

    try {
      // Start OAuth server
      const serverResult = await window.electronAPI.startOAuthServer();
      if (!serverResult.success) {
        throw new Error(serverResult.error || 'Failed to start OAuth server');
      }

      // Get OAuth URL from backend
      let urlResult;
      if (provider === 'google') {
        urlResult = await AuthService.getGoogleOAuthURL();
      } else {
        // For GitHub, we need to pass the JWT token
        const currentToken = storageService.getToken();
        if (!currentToken) {
          throw new Error('Must be authenticated to add GitHub storage accounts');
        }
        urlResult = await AuthService.getGitHubOAuthURL(currentToken);
      }

      if (!urlResult.success || !urlResult.authURL) {
        throw new Error('Failed to get OAuth URL from backend');
      }

      // Open OAuth URL in browser
      await window.electronAPI.openExternal(urlResult.authURL);

      toast({
        title: 'Authentication Started',
        description: 'Browser opened. Please complete the authentication process.',
        status: 'info',
        duration: 3000,
        isClosable: true,
      });

    } catch (error) {
      setIsLoading(false);
      toast({
        title: 'Failed to Start Authentication',
        description: (error as Error).message,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  const logout = () => {
    storageService.clearAuth();
    setAuthState(storageService.getAuthState());
    toast({
      title: 'Logged Out',
      description: 'Successfully logged out',
      status: 'info',
      duration: 2000,
      isClosable: true,
    });
  };

  const updateBackendURL = () => {
    storageService.setBackendURL(backendURL);
    AuthService.setBackendURL(backendURL);
    checkServerConnection();
    toast({
      title: 'Backend URL Updated',
      description: 'Checking connection...',
      status: 'info',
      duration: 2000,
      isClosable: true,
    });
  };

  return (
    <ChakraProvider>
      <Box p={6} minH="100vh" bg="gray.50">
        <VStack spacing={6} maxW="md" mx="auto">
          <Heading size="lg" color="gray.700">
            DataHub OAuth Test
          </Heading>

          {/* Server Status */}
          <Card w="full">
            <CardBody>
              <VStack spacing={3} align="stretch">
                <HStack justify="space-between">
                  <Text fontWeight="semibold">Server Status:</Text>
                  <Badge colorScheme={serverOnline ? 'green' : 'red'}>
                    {serverOnline === null ? 'Checking...' : serverOnline ? 'Online' : 'Offline'}
                  </Badge>
                </HStack>
                
                <FormControl>
                  <FormLabel fontSize="sm">Backend URL:</FormLabel>
                  <HStack>
                    <Input
                      value={backendURL}
                      onChange={(e) => setBackendURL(e.target.value)}
                      placeholder="http://localhost:8080"
                      size="sm"
                    />
                    <Button size="sm" onClick={updateBackendURL}>
                      Update
                    </Button>
                  </HStack>
                </FormControl>
              </VStack>
            </CardBody>
          </Card>

          {/* Authentication Status */}
          <Card w="full">
            <CardBody>
              <VStack spacing={4} align="stretch">
                <HStack justify="space-between">
                  <Text fontWeight="semibold">Authentication Status:</Text>
                  <Badge colorScheme={authState.isAuthenticated ? 'green' : 'gray'}>
                    {authState.isAuthenticated ? 'Authenticated' : 'Not Authenticated'}
                  </Badge>
                </HStack>

                {authState.isAuthenticated && authState.user && (
                  <VStack align="stretch" spacing={2}>
                    <Text fontSize="sm">
                      <strong>Email:</strong> {authState.user.email || 'N/A'}
                    </Text>
                    <Text fontSize="sm">
                      <strong>Name:</strong> {authState.user.name || 'N/A'}
                    </Text>
                  </VStack>
                )}

                {authState.connectedGitHubAccounts.length > 0 && (
                  <VStack align="stretch" spacing={2}>
                    <Text fontSize="sm" fontWeight="semibold">Connected GitHub Accounts:</Text>
                    {authState.connectedGitHubAccounts.map((account) => (
                      <Badge key={account} colorScheme="purple" variant="subtle">
                        {account}
                      </Badge>
                    ))}
                  </VStack>
                )}
              </VStack>
            </CardBody>
          </Card>

          {/* OAuth Actions */}
          <Card w="full">
            <CardBody>
              <VStack spacing={4}>
                <Text fontWeight="semibold">OAuth Authentication:</Text>
                
                {!authState.isAuthenticated ? (
                  <VStack spacing={3} w="full">
                    <Button
                      colorScheme="red"
                      onClick={() => startOAuth('google')}
                      isLoading={isLoading}
                      loadingText="Waiting for auth..."
                      w="full"
                      disabled={!serverOnline}
                    >
                      {isLoading ? <Spinner size="sm" mr={2} /> : null}
                      Sign in with Google
                    </Button>
                    
                    <Text fontSize="sm" color="gray.500" textAlign="center" mt={2}>
                      GitHub accounts can be added after signing in with Google
                    </Text>
                  </VStack>
                ) : (
                  <VStack spacing={3} w="full">
                    <Button
                      colorScheme="gray"
                      onClick={() => startOAuth('github')}
                      isLoading={isLoading}
                      loadingText="Waiting for auth..."
                      w="full"
                      disabled={!serverOnline}
                    >
                      Connect GitHub Storage Account
                    </Button>
                    
                    <Divider />
                    
                    <Button
                      colorScheme="red"
                      variant="outline"
                      onClick={logout}
                      w="full"
                    >
                      Logout
                    </Button>
                  </VStack>
                )}
              </VStack>
            </CardBody>
          </Card>

          {/* Instructions */}
          {!serverOnline && (
            <Alert status="warning">
              <AlertIcon />
              <Box>
                <AlertTitle>Backend Not Running!</AlertTitle>
                <AlertDescription>
                  Please start your Go backend server before testing OAuth flows.
                </AlertDescription>
              </Box>
            </Alert>
          )}
        </VStack>
      </Box>
    </ChakraProvider>
  );
};

export default App;