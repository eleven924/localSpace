export namespace models {
	
	export class AIAnalysis {
	    tags: string[];
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new AIAnalysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tags = source["tags"];
	        this.description = source["description"];
	    }
	}
	export class AIConfig {
	    id: number;
	    apiKey: string;
	    model: string;
	    baseURL: string;
	    enabled: boolean;
	    enableAgent: boolean;
	    enableWebSearch: boolean;
	    maxTokens: number;
	    timeout: number;
	    webSearchProvider: string;
	    webSearchBaseURL: string;
	    webSearchAPIKey: string;
	    webSearchTimeout: number;
	    webSearchMaxResults: number;
	
	    static createFrom(source: any = {}) {
	        return new AIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.apiKey = source["apiKey"];
	        this.model = source["model"];
	        this.baseURL = source["baseURL"];
	        this.enabled = source["enabled"];
	        this.enableAgent = source["enableAgent"];
	        this.enableWebSearch = source["enableWebSearch"];
	        this.maxTokens = source["maxTokens"];
	        this.timeout = source["timeout"];
	        this.webSearchProvider = source["webSearchProvider"];
	        this.webSearchBaseURL = source["webSearchBaseURL"];
	        this.webSearchAPIKey = source["webSearchAPIKey"];
	        this.webSearchTimeout = source["webSearchTimeout"];
	        this.webSearchMaxResults = source["webSearchMaxResults"];
	    }
	}
	export class BatchDeleteFailedItem {
	    fileId: number;
	    fileName: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new BatchDeleteFailedItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fileId = source["fileId"];
	        this.fileName = source["fileName"];
	        this.error = source["error"];
	    }
	}
	export class BatchDeleteResult {
	    successCount: number;
	    failedCount: number;
	    failedItems: BatchDeleteFailedItem[];
	
	    static createFrom(source: any = {}) {
	        return new BatchDeleteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.successCount = source["successCount"];
	        this.failedCount = source["failedCount"];
	        this.failedItems = this.convertValues(source["failedItems"], BatchDeleteFailedItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BatchImportFileInput {
	    sourcePath: string;
	    displayName: string;
	
	    static createFrom(source: any = {}) {
	        return new BatchImportFileInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourcePath = source["sourcePath"];
	        this.displayName = source["displayName"];
	    }
	}
	export class BatchImportJobRequest {
	    files: BatchImportFileInput[];
	    sharedTags: string[];
	    sharedDescription: string;
	    collectionId?: number;
	    enableAIGeneratedTags: boolean;
	    enableAIGeneratedDescription: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BatchImportJobRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.files = this.convertValues(source["files"], BatchImportFileInput);
	        this.sharedTags = source["sharedTags"];
	        this.sharedDescription = source["sharedDescription"];
	        this.collectionId = source["collectionId"];
	        this.enableAIGeneratedTags = source["enableAIGeneratedTags"];
	        this.enableAIGeneratedDescription = source["enableAIGeneratedDescription"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BatchMoveFailedItem {
	    fileId: number;
	    fileName: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new BatchMoveFailedItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fileId = source["fileId"];
	        this.fileName = source["fileName"];
	        this.error = source["error"];
	    }
	}
	export class BatchMoveResult {
	    successCount: number;
	    failedCount: number;
	    failedItems: BatchMoveFailedItem[];
	
	    static createFrom(source: any = {}) {
	        return new BatchMoveResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.successCount = source["successCount"];
	        this.failedCount = source["failedCount"];
	        this.failedItems = this.convertValues(source["failedItems"], BatchMoveFailedItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Collection {
	    id: number;
	    name: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Collection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class CollectionFilterCounts {
	    total: number;
	    unsorted: number;
	    collections: Record<number, number>;
	
	    static createFrom(source: any = {}) {
	        return new CollectionFilterCounts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.unsorted = source["unsorted"];
	        this.collections = source["collections"];
	    }
	}
	export class Metadata {
	    size?: number;
	    width?: number;
	    height?: number;
	    duration?: number;
	    pageCount?: number;
	    author?: string;
	    title?: string;
	
	    static createFrom(source: any = {}) {
	        return new Metadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.size = source["size"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.duration = source["duration"];
	        this.pageCount = source["pageCount"];
	        this.author = source["author"];
	        this.title = source["title"];
	    }
	}
	export class File {
	    id: number;
	    fileName: string;
	    originalName: string;
	    collectionName: string;
	    collectionId?: number;
	    filePath: string;
	    fileType: string;
	    fileSubType: string;
	    fileSize: number;
	    tags: string[];
	    description: string;
	    metadata: Metadata;
	    thumbnail: string;
	    checksum: string;
	    isDeleted: boolean;
	    deletedAt: string;
	    createdAt: string;
	    modifiedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new File(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fileName = source["fileName"];
	        this.originalName = source["originalName"];
	        this.collectionName = source["collectionName"];
	        this.collectionId = source["collectionId"];
	        this.filePath = source["filePath"];
	        this.fileType = source["fileType"];
	        this.fileSubType = source["fileSubType"];
	        this.fileSize = source["fileSize"];
	        this.tags = source["tags"];
	        this.description = source["description"];
	        this.metadata = this.convertValues(source["metadata"], Metadata);
	        this.thumbnail = source["thumbnail"];
	        this.checksum = source["checksum"];
	        this.isDeleted = source["isDeleted"];
	        this.deletedAt = source["deletedAt"];
	        this.createdAt = source["createdAt"];
	        this.modifiedAt = source["modifiedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileListResponse {
	    items: File[];
	    page: number;
	    pageSize: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new FileListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], File);
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileType {
	    id: number;
	    name: string;
	    displayName: string;
	    extensions: string;
	    subTypes: string[];
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new FileType(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.displayName = source["displayName"];
	        this.extensions = source["extensions"];
	        this.subTypes = source["subTypes"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class Job {
	    id: number;
	    jobType: string;
	    status: string;
	    title: string;
	    payload: number[];
	    result: number[];
	    progressTotal: number;
	    progressCompleted: number;
	    progressMessage: string;
	    exclusiveKey: string;
	    canResume: boolean;
	    startedAt: string;
	    heartbeatAt: string;
	    finishedAt: string;
	    timeoutAt: string;
	    errorMessage: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Job(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.jobType = source["jobType"];
	        this.status = source["status"];
	        this.title = source["title"];
	        this.payload = source["payload"];
	        this.result = source["result"];
	        this.progressTotal = source["progressTotal"];
	        this.progressCompleted = source["progressCompleted"];
	        this.progressMessage = source["progressMessage"];
	        this.exclusiveKey = source["exclusiveKey"];
	        this.canResume = source["canResume"];
	        this.startedAt = source["startedAt"];
	        this.heartbeatAt = source["heartbeatAt"];
	        this.finishedAt = source["finishedAt"];
	        this.timeoutAt = source["timeoutAt"];
	        this.errorMessage = source["errorMessage"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class JobListResponse {
	    items: Job[];
	    page: number;
	    pageSize: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new JobListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], Job);
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class JobRetentionConfig {
	    id: number;
	    enabled: boolean;
	    maxCount: number;
	    maxDays: number;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new JobRetentionConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.enabled = source["enabled"];
	        this.maxCount = source["maxCount"];
	        this.maxDays = source["maxDays"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	
	export class OpenWithConfig {
	    byFileType: Record<string, string>;
	    byExtension: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new OpenWithConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.byFileType = source["byFileType"];
	        this.byExtension = source["byExtension"];
	    }
	}
	export class SelectedFile {
	    name: string;
	    path: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new SelectedFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	    }
	}
	export class StorageDir {
	    id: number;
	    path: string;
	    fileType: string;
	    currentSize: number;
	    maxSize: number;
	    isActive: boolean;
	    isDefault: boolean;
	    parentId?: number;
	    createdAt: string;
	    totalSize?: number;
	    subDirs?: StorageDir[];
	
	    static createFrom(source: any = {}) {
	        return new StorageDir(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.fileType = source["fileType"];
	        this.currentSize = source["currentSize"];
	        this.maxSize = source["maxSize"];
	        this.isActive = source["isActive"];
	        this.isDefault = source["isDefault"];
	        this.parentId = source["parentId"];
	        this.createdAt = source["createdAt"];
	        this.totalSize = source["totalSize"];
	        this.subDirs = this.convertValues(source["subDirs"], StorageDir);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StorageLayoutConfig {
	    strategy: string;
	    unsortedFolderName: string;
	    sanitizeFolderName: boolean;
	
	    static createFrom(source: any = {}) {
	        return new StorageLayoutConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.strategy = source["strategy"];
	        this.unsortedFolderName = source["unsortedFolderName"];
	        this.sanitizeFolderName = source["sanitizeFolderName"];
	    }
	}
	export class ThemeConfig {
	    id: number;
	    themeMode: string;
	    primaryColor: string;
	    backgroundColor: string;
	    backgroundImage: string;
	
	    static createFrom(source: any = {}) {
	        return new ThemeConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.themeMode = source["themeMode"];
	        this.primaryColor = source["primaryColor"];
	        this.backgroundColor = source["backgroundColor"];
	        this.backgroundImage = source["backgroundImage"];
	    }
	}

}

