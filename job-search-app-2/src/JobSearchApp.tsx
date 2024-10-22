import React, { useState } from 'react';
import { Input } from './components/ui/input';
import { Button } from './components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from './components/ui/card';
import { Label } from './components/ui/label';
import { Alert, AlertDescription } from './components/ui/alert';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "./components/ui/collapsible"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./components/ui/tabs"
import { RadioGroup, RadioGroupItem } from './components/ui/radio-group';
import { ChevronDown, ChevronUp, Upload, Linkedin, ExternalLink, Star } from 'lucide-react';


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
  const [resumeFile, setResumeFile] = useState(null);
  const [linkedInUrl, setLinkedInUrl] = useState('');
  const [fileError, setFileError] = useState('');
  const [isAdvancedOpen, setIsAdvancedOpen] = useState(false);
  const [profileOption, setProfileOption] = useState('resume');
  const [searchDetails, setSearchDetails] = useState<SearchDetails | null>(null);
  const [isSearchDetailsOpen, setIsSearchDetailsOpen] = useState(false);

  const searchJobs = async (searchType: string) => {
    setLoading(true);
    try {
      const formData = new FormData();
      formData.append('months', months.toString());
      formData.append('prompt', prompt);
      formData.append('type', searchType);
      if (profileOption === 'linkedin') {
        formData.append('linkedin', linkedInUrl);
      } else {
        if (resumeFile) {
          formData.append('file', resumeFile);
        }
      }
      const response = await fetch('http://localhost:8080/jobs', {
        method: 'POST',
        body: formData,
      });
      const data = await response.json();
      const items = data.items || [];
      setResults((data.comments || []).map((id: number) => items.find((item: Item) => item.id === id)));
      setParents((data.parents || []).map((id: number) => items.find((item: Item) => item.id === id)));
      setSearchDetails({
        hackerNewsLinks: data.hacker_news_links || [],
        originalHackerNewsLinks: data.original_hacker_news_links || [],
        resumeSummary: data.resume_summary || "",
        searchTerms: data.search_terms || [],
        totalPosts: data.total_posts || 0,
        totalItems: data.total_items || 0,
        posts: data.posts || 0,
        itemsSearched: data.items_searched || 0,
        latencies: data.latencies || {},
      });
    } catch (error) {
      console.error('Error fetching jobs:', error);
      setResults([]);
    }
    setLoading(false);
  };

  const handleResumeUpload = (event : any) => {
    const selectedFile = event.target.files[0];
    if (selectedFile) {
      if (selectedFile.type === 'application/pdf' || selectedFile.type === 'text/plain') {
        setResumeFile(selectedFile);
        setFileError('');
      } else {
        setResumeFile(null);
        setFileError('Please upload only PDF or text files.');
      }
    }
  };

  const handleLinkedInUrlChange = (event : any) => {
    setLinkedInUrl(event.target.value);
  };

  const renderSearchSection = (type: any) => (
      <div className="space-y-4">
        <div>
          <CardContent className="pt-6">
            <div className="mt-6">
              <Label className="text-lg font-semibold mb-2 block">Choose your profile option:</Label>
              <RadioGroup
                  value={profileOption}
                  onValueChange={setProfileOption}
                  className="flex flex-col space-y-2"
              >
                <div className="flex items-center space-x-2">
                  <RadioGroupItem value="resume" id="resume-option" />
                  <Label htmlFor="resume-option">Upload Resume</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <RadioGroupItem value="linkedin" id="linkedin-option" />
                  <Label htmlFor="linkedin-option">Add LinkedIn Profile</Label>
                </div>
              </RadioGroup>
            </div>
            {profileOption === 'resume' && (
                <div className="mt-4">
                  <Label htmlFor="resume-upload" className="block mb-2">Upload Resume</Label>
                  <div className="flex items-center space-x-2">
                    <Input
                        id="resume-upload"
                        type="file"
                        accept=".pdf,.txt"
                        onChange={handleResumeUpload}
                        className="file:mr-4 file:py-1 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-hn-orange file:text-white hover:file:bg-hn-orange-dark"
                    />
                    <Upload className="text-hn-orange" />
                  </div>
                  {resumeFile && <p className="mt-2 text-sm text-hn-text"> File uploaded
                  </p>}
                  {fileError && (
                      <Alert variant="destructive" className="mt-2">
                        <AlertDescription>{fileError}</AlertDescription>
                      </Alert>
                  )}
                </div>
            )}
            {profileOption === 'linkedin' && (
                <div className="mt-4">
                  <Label htmlFor="linkedin-url" className="block mb-2">LinkedIn Profile URL</Label>
                  <div className="flex items-center space-x-2">
                    <Input
                        id="linkedin-url"
                        type="url"
                        placeholder="https://www.linkedin.com/in/yourprofile"
                        value={linkedInUrl}
                        onChange={handleLinkedInUrlChange}
                        className="flex-grow"
                    />
                    <Linkedin className="text-hn-orange" />
                  </div>
                </div>
            )}
          </CardContent>
        </div>
        <Collapsible open={isAdvancedOpen} onOpenChange={setIsAdvancedOpen}>
          <CollapsibleTrigger asChild>
            <Button variant="outline" className="w-full">
              Advanced Options {isAdvancedOpen ? <ChevronUp className="ml-2 h-4 w-4"/> :
                <ChevronDown className="ml-2 h-4 w-4"/>}
            </Button>
          </CollapsibleTrigger>
          <CollapsibleContent className="space-y-4 mt-4">
            <div>
              <Label htmlFor="months">Number of months to search</Label>
              <Input
                  id="months"
                  type="number"
                  value={months}
                  // @ts-ignore
                  onChange={(e) => setMonths(e.target.value)}
                  min="1"
              />
            </div>
            <div>
              <Label htmlFor="prompt">Search prompt</Label>
              <Input
                  id="prompt"
                  value={prompt}
                  onChange={(e) => setPrompt(e.target.value)}
                  placeholder="Enter additional search criteria"
              />
            </div>
          </CollapsibleContent>
        </Collapsible>
        <Button onClick={() => searchJobs(type)} disabled={loading} className="w-full">
          {loading ? 'Searching...' : type === 'hiring' ? 'Find a New Job' : 'Find me a rockstar!'}
          {type !== 'hiring' && (
          <Star className="text-hn-orange" />
          )}
        </Button>
      </div>
  );

  interface SearchDetails {
    hackerNewsLinks: string[];
    originalHackerNewsLinks: string[];
    resumeSummary: string;
    searchTerms: string[];
    totalItems: number;
    totalPosts: number;
    posts: number;
    itemsSearched: number;
    latencies: {[key: string]: number};
  }

  const SearchDetailsSection = () => {
    if (!searchDetails) return null;

    return (
        <Collapsible
            open={isSearchDetailsOpen}
            onOpenChange={setIsSearchDetailsOpen}
            className="mt-8 space-y-4"
        >
          <CollapsibleTrigger asChild>
            <Button
                variant="outline"
                className="flex items-center justify-between w-full"
            >
              <span>Search Details</span>
              {isSearchDetailsOpen ? (
                  <ChevronUp className="h-4 w-4" />
              ) : (
                  <ChevronDown className="h-4 w-4" />
              )}
            </Button>
          </CollapsibleTrigger>
          <CollapsibleContent className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Hacker News Links</CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="list-disc list-inside">
                  {searchDetails.hackerNewsLinks.map((link, index) => (
                      <li key={index} className="flex items-center">
                        <a href={link} target="_blank" rel="noopener noreferrer" className="text-hn-orange hover:underline">
                          {link}
                        </a>
                        <ExternalLink className="ml-2 h-4 w-4 text-hn-text" />
                      </li>
                  ))}
                </ul>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Pre-Search Hacker News Links</CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="list-disc list-inside">
                  {searchDetails.originalHackerNewsLinks.map((link, index) => (
                      <li key={index} className="flex items-center">
                        <a href={link} target="_blank" rel="noopener noreferrer" className="text-hn-orange hover:underline">
                          {link}
                        </a>
                        <ExternalLink className="ml-2 h-4 w-4 text-hn-text" />
                      </li>
                  ))}
                </ul>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Resume Summary</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-hn-text whitespace-pre-wrap">{searchDetails.resumeSummary}</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Search Terms</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2">
                  {searchDetails.searchTerms.map((keyword, index) => (
                      <span key={index} className="bg-hn-text text-white px-2 py-1 rounded-full text-sm">
                    {keyword}
                  </span>
                  ))}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Search Statistics</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-1">
                  {Object.entries(searchDetails.latencies).map(([key, value]) => (
                      <div key={key} className="text-hn-text">
                        <span className="capitalize">{key.replace(/_/g, ' ')}: </span>
                        <span className="font-mono text-hn-orange">{value.toFixed(6)} s</span>
                      </div>
                  ))}
                  <p className="text-hn-text">
                    Total Items: <span className="font-bold">{searchDetails.totalItems}</span>
                  </p>
                  <p className="text-hn-text">
                    Total Posts: <span className="font-bold">{searchDetails.totalPosts}</span>
                  </p>
                  <p className="text-hn-text">
                    Posts Searched: <span className="font-bold">{searchDetails.posts}</span>
                  </p>
                  <p className="text-hn-text">
                    Items Searched: <span className="font-bold">{searchDetails.itemsSearched}</span>
                  </p>
                </div>
              </CardContent>
            </Card>
          </CollapsibleContent>
        </Collapsible>
    );
  };

  return (
      <div className="p-4 max-w-6xl mx-auto bg-hn-background min-h-screen font-sans text-hn-base">
        <h1 className="text-hn-large font-bold mb-8 text-center text-hn-orange">Job Search Application</h1>
        <Tabs defaultValue="hiring" className="w-full">
          <TabsList className="grid w-full grid-cols-2 bg-hn-orange">
            <TabsTrigger value="hiring"
                         className="text-white data-[state=active]:bg-white data-[state=active]:text-hn-orange">Who is
              Hiring</TabsTrigger>
            <TabsTrigger value="seekers"
                         className="text-white data-[state=active]:bg-white data-[state=active]:text-hn-orange">Who
              wants to be Hired</TabsTrigger>
          </TabsList>
          <TabsContent value="hiring">
            <Card className="border-hn-orange">
              <CardContent className="pt-6">
                {renderSearchSection('hiring')}
              </CardContent>
            </Card>
          </TabsContent>
          <TabsContent value="seekers">
            <Card className="border-hn-orange">
              <CardContent className="pt-6">
                {renderSearchSection('seekers')}
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
        {searchDetails && <SearchDetailsSection />}
        <div className="mt-8">
          <h2 className="text-xl font-semibold mb-4 text-hn-text">Search Results</h2>
          {results.map((item, index) => (
              <Card key={index} className="mb-4">
                <CardHeader>
                  <CardTitleComponent item={item}/>
                </CardHeader>
                <CardContent dangerouslySetInnerHTML={{__html: item.text}}/>
              </Card>
          ))}
          {results.length === 0 && !loading && (
              <p className="text-hn-text">No results found. Try adjusting your search parameters.</p>
          )}
        </div>
      </div>
  );
};

export default JobSearchApp;