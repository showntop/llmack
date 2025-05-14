import React, { useState, useRef, useEffect } from 'react';
import {
  Box,
  TextField,
  Button,
  Paper,
  Typography,
  CircularProgress,
  useMediaQuery,
  useTheme,
  IconButton,
} from '@mui/material';
import SendIcon from '@mui/icons-material/Send';
import AttachFileIcon from '@mui/icons-material/AttachFile';
import axios from 'axios';

interface Message {
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
}

interface ChatInterfaceProps {
  token: string;
}

export const ChatInterface: React.FC<ChatInterfaceProps> = ({ token }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = async () => {
    if (!input.trim() && !selectedFile) return;

    const newMessage: Message = {
      role: 'user',
      content: input,
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, newMessage]);
    setInput('');
    setIsLoading(true);

    try {
      if (selectedFile) {
        const formData = new FormData();
        formData.append('file', selectedFile);

        const response = await axios.post('http://localhost:8080/api/transactions/upload', formData, {
          headers: {
            'Authorization': token,
            'Content-Type': 'multipart/form-data',
          },
        });

        setMessages((prev) => [
          ...prev,
          {
            role: 'assistant',
            content: `已成功上传并分析图片：${response.data.message}`,
            timestamp: new Date(),
          },
        ]);
      } else {
        const response = await axios.post(
          '/api/chat',
          { message: input },
          {
            headers: {
              'Authorization': token,
            },
          }
        );

        setMessages((prev) => [
          ...prev,
          {
            role: 'assistant',
            content: response.data.response,
            timestamp: new Date(),
          },
        ]);
      }
    } catch (error) {
      console.error('Error:', error);
      setMessages((prev) => [
        ...prev,
        {
          role: 'assistant',
          content: '抱歉，处理您的请求时出现错误。',
          timestamp: new Date(),
        },
      ]);
    } finally {
      setIsLoading(false);
      setSelectedFile(null);
    }
  };

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setSelectedFile(file);
    }
  };

  const handleKeyPress = (event: React.KeyboardEvent) => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      handleSend();
    }
  };

  return (
    <Box
      sx={{
        height: 'calc(100vh - 180px)',
        display: 'flex',
        flexDirection: 'column',
        gap: 2,
      }}
    >
      <Paper
        elevation={3}
        sx={{
          flex: 1,
          overflow: 'auto',
          p: 2,
          display: 'flex',
          flexDirection: 'column',
          gap: 2,
        }}
      >
        {messages.map((message, index) => (
          <Box
            key={index}
            sx={{
              display: 'flex',
              justifyContent: message.role === 'user' ? 'flex-end' : 'flex-start',
            }}
          >
            <Paper
              elevation={1}
              sx={{
                p: 2,
                maxWidth: isMobile ? '80%' : '60%',
                backgroundColor: message.role === 'user' ? 'primary.light' : 'grey.100',
                color: message.role === 'user' ? 'white' : 'text.primary',
              }}
            >
              <Typography variant="body1">{message.content}</Typography>
              <Typography variant="caption" sx={{ display: 'block', mt: 1, opacity: 0.7 }}>
                {message.timestamp.toLocaleTimeString()}
              </Typography>
            </Paper>
          </Box>
        ))}
        {isLoading && (
          <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
            <CircularProgress size={24} />
          </Box>
        )}
        <div ref={messagesEndRef} />
      </Paper>

      <Box sx={{ display: 'flex', gap: 1, alignItems: 'flex-end' }}>
        <input
          type="file"
          accept="image/*"
          style={{ display: 'none' }}
          ref={fileInputRef}
          onChange={handleFileSelect}
        />
        <IconButton
          color="primary"
          onClick={() => fileInputRef.current?.click()}
          sx={{ alignSelf: 'flex-end' }}
        >
          <AttachFileIcon />
        </IconButton>
        <TextField
          fullWidth
          multiline
          maxRows={4}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyPress={handleKeyPress}
          placeholder="输入消息..."
          variant="outlined"
          size="small"
        />
        <Button
          variant="contained"
          color="primary"
          onClick={handleSend}
          disabled={isLoading || (!input.trim() && !selectedFile)}
          sx={{ minWidth: 'auto', px: 2 }}
        >
          <SendIcon />
        </Button>
      </Box>
      {selectedFile && (
        <Typography variant="caption" color="primary">
          已选择文件: {selectedFile.name}
        </Typography>
      )}
    </Box>
  );
}; 