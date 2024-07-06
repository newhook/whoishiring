import React, { useState } from 'react';
import { Input } from './components/ui/input';
import { Button } from './components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from './components/ui/card';
import { RadioGroup, RadioGroupItem } from './components/ui/radio-group';
import { Label } from './components/ui/label';

interface Item {
  id: number;
  type: string;
  by: string;
  time: number;
  text: string;
  parent: number;
  title: string;
  descendants: number;
  kids: number[];
  score: number;
}

// Helper function to calculate the difference in days
const getTimeDifferenceInDays = (pastUnixTimestamp: number): number => {
  const pastDate = new Date(pastUnixTimestamp * 1000);
  const currentDate = new Date();
  const differenceInTime = currentDate.getTime() - pastDate.getTime();
  const differenceInDays = Math.floor(differenceInTime / (1000 * 3600 * 24));
  return differenceInDays;
};

// Component to render the CardTitle
const CardTitleComponent = ({ item }: { item: Item }) => {
  const daysAgo = getTimeDifferenceInDays(item.time);
  const title = `${item.by} ${daysAgo} days ago | parent | context | favorite | on: ${item.title}`;

  return (
      <CardTitle>
        <a href={`https://news.ycombinator.com/item?id=${item.id}`} target="_blank" rel="noopener noreferrer">
          {title}
        </a>
      </CardTitle>
  );
};

const JobSearchApp = () => {
  const [months, setMonths] = useState(1);
  const [prompt, setPrompt] = useState('');
  const [results, setResults] = useState<Item[]>([]);
  const [parents, setParents] = useState<Item[]>([]);
  const [searchType, setSearchType] = useState('hiring');
  const [loading, setLoading] = useState(false);

  const searchJobs = async () => {
    setLoading(true);
    try {
      // Replace this URL with your actual API endpoint
      // const response = await fetch(`https://api.example.com/jobs?months=${months}&prompt=${encodeURIComponent(prompt)}`);
      const response = await fetch(`http://localhost:8080/jobs?months=${months}&prompt=${encodeURIComponent(prompt)}&type=${searchType}`);
      const data = await response.json();
      setResults(data.comments || []);
      setParents(data.parents || []);
    } catch (error) {
      console.error('Error fetching jobs:', error);
      setResults([]);
    }
    setLoading(false);
  };

  return (
    <div className="p-4 max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-4">Job Search Application</h1>
      <div className="space-y-4">
        <Input
          type="number"
          value={months}
          // @ts-ignore
          onChange={(e) => setMonths(e.target.value)}
          placeholder="Number of months to search"
          min="1"
        />
        <Input
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          placeholder="Search prompt"
        />
        <RadioGroup defaultValue="hiring" onValueChange={setSearchType}>
          <div className="flex items-center space-x-2">
            <RadioGroupItem value="hiring" id="hiring" />
            <Label htmlFor="hiring">Who is hiring</Label>
          </div>
          <div className="flex items-center space-x-2">
            <RadioGroupItem value="seekers" id="seekers" />
            <Label htmlFor="seekers">Who wants to be hired</Label>
          </div>
        </RadioGroup>
        <Button onClick={searchJobs} disabled={loading}>
          {loading ? 'Searching...' : 'Search Jobs'}
        </Button>
      </div>
      <div className="mt-8">
        <h2 className="text-xl font-semibold mb-4">Search Results</h2>
        {results.map((item, index) => (
          <Card key={index} className="mb-4">
            <CardHeader>
              <CardTitleComponent item={item} />
            </CardHeader>
            <CardContent dangerouslySetInnerHTML={{ __html: item.text }} />
          </Card>
        ))}
        {results.length === 0 && !loading && (
          <p>No results found. Try adjusting your search parameters.</p>
        )}
      </div>
    </div>
  );
};

export default JobSearchApp;