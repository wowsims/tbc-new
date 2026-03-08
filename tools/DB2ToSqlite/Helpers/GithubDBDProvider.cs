using DBCD.Providers;

namespace DB2ToSqliteTool.Helpers;

public class CustomGithubDBDProvider : IDBDProvider
{
	private static Uri BaseURI = new("https://raw.githubusercontent.com/wowdev/WoWDBDefs/master/definitions/");
	private readonly HttpClient client = new();

	private static bool UseCache = false;
	private static string CachePath { get; } = "DBDCache/";
	private static readonly TimeSpan CacheExpiryTime = new(1, 0, 0, 0);

	public CustomGithubDBDProvider(bool useCache = false, string? customUri = null)
	{
		if (!string.IsNullOrEmpty(customUri))
		{
			BaseURI = new Uri(customUri);
		}

		UseCache = useCache;
		if (useCache && !Directory.Exists(CachePath))
			Directory.CreateDirectory(CachePath);

		client.BaseAddress = BaseURI;
		Console.WriteLine($"Initialized CustomGithubDBDProvider with base URI: {client.BaseAddress}");
	}

	public Stream StreamForTableName(string tableName, string build)
	{
		var query = $"{tableName}.dbd";

		if (UseCache)
		{
			var cacheFile = Path.Combine(CachePath, query);
			if (File.Exists(cacheFile))
			{
				var lastWrite = File.GetLastWriteTime(cacheFile);
				if (DateTime.Now - lastWrite < CacheExpiryTime)
					return new MemoryStream(File.ReadAllBytes(cacheFile));
			}
		}

		var bytes = client.GetByteArrayAsync(query).Result;

		if (UseCache)
			File.WriteAllBytes(Path.Combine(CachePath, query), bytes);

		return new MemoryStream(bytes);
	}
}
