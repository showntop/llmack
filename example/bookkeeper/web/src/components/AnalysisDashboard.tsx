import React, { useState, useEffect } from 'react';
import {
  Box,
  Grid,
  Paper,
  Typography,
  CircularProgress,
  useTheme,
  useMediaQuery,
  Card,
  CardContent,
} from '@mui/material';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts';
import axios from 'axios';

interface AnalysisDashboardProps {
  token: string;
}

interface AnalysisData {
  total_expenses: number;
  total_income: number;
  category_distribution: Array<{
    category: string;
    amount: number;
  }>;
  monthly_trend: Array<{
    month: string;
    expenses: number;
    income: number;
  }>;
}

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8'];

export const AnalysisDashboard: React.FC<AnalysisDashboardProps> = ({ token }) => {
  const [data, setData] = useState<AnalysisData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  useEffect(() => {
    fetchAnalysisData();
  }, [token]);

  const fetchAnalysisData = async () => {
    try {
      const response = await axios.get('/api/analysis/monthly', {
        headers: {
          Authorization: token,
        },
      });
      setData(response.data);
    } catch (err) {
      setError('获取分析数据失败');
      console.error('Error fetching analysis data:', err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box p={3}>
        <Typography color="error">{error}</Typography>
      </Box>
    );
  }

  if (!data) {
    return null;
  }

  return (
    <Box p={isMobile ? 1 : 3}>
      <Grid container spacing={isMobile ? 1 : 3}>
        <Grid item xs={12} sm={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                总支出
              </Typography>
              <Typography variant="h4" color="error">
                ¥{data.total_expenses.toFixed(2)}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                总收入
              </Typography>
              <Typography variant="h4" color="success.main">
                ¥{data.total_income.toFixed(2)}
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Paper sx={{ p: isMobile ? 1 : 2, height: '100%' }}>
            <Typography variant="h6" gutterBottom>
              支出分类分布
            </Typography>
            <Box sx={{ height: 300 }}>
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={data.category_distribution}
                    dataKey="amount"
                    nameKey="category"
                    cx="50%"
                    cy="50%"
                    outerRadius={80}
                    label
                  >
                    {data.category_distribution.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            </Box>
          </Paper>
        </Grid>

        <Grid item xs={12} md={6}>
          <Paper sx={{ p: isMobile ? 1 : 2, height: '100%' }}>
            <Typography variant="h6" gutterBottom>
              月度收支趋势
            </Typography>
            <Box sx={{ height: 300 }}>
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={data.monthly_trend}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="month" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Bar dataKey="expenses" name="支出" fill="#FF8042" />
                  <Bar dataKey="income" name="收入" fill="#00C49F" />
                </BarChart>
              </ResponsiveContainer>
            </Box>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
}; 